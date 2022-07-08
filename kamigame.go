package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"golang.org/x/exp/slices"
)

var _ provider = &kamigame{}

// https://kamigame.jp/umamusume/page/152540608660042049.html
type kamigame struct {
	character, support, skills [][]any
	data                       map[string]string
}

func (p kamigame) name() string { return "kamigame" }

func (p *kamigame) events(process bool) (events []event, err error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()

	var ids []network.RequestID
	done := make(chan network.RequestID, 4)
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			if strings.HasPrefix(ev.Request.URL, "https://kamigame.jp/vls-kamigame-gametool/json") {
				ids = append(ids, ev.RequestID)
			}
		case *network.EventLoadingFinished:
			if slices.Contains(ids, ev.RequestID) {
				done <- ev.RequestID
			}
		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(ctx)
				ctx := cdp.WithExecutor(ctx, c.Target)

				if ev.ResourceType == network.ResourceTypeDocument ||
					(ev.ResourceType == network.ResourceTypeScript && !strings.Contains(ev.Request.URL, "ad")) ||
					ev.ResourceType == network.ResourceTypeXHR {
					//log.Println("allow:", ev.Request.URL)
					fetch.ContinueRequest(ev.RequestID).Do(ctx)
				} else {
					// log.Println("block:", ev.Request.URL)
					fetch.FailRequest(ev.RequestID, network.ErrorReasonBlockedByClient).Do(ctx)
				}
			}()
		}
	})

	if err = chromedp.Run(
		ctx,
		fetch.Enable(),
		chromedp.Navigate("https://kamigame.jp/umamusume/page/152540608660042049.html"),
	); err != nil {
		return
	}

	var res [][]byte
	for i := 0; i < 4; i++ {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case id := <-done:
			if err = chromedp.Run(
				ctx,
				chromedp.ActionFunc(func(ctx context.Context) error {
					b, err := network.GetResponseBody(id).Do(ctx)
					if err != nil {
						return err
					}
					res = append(res, b)
					return nil
				}),
			); err != nil {
				return
			}
		}
	}

	var choices [][]any
	for _, b := range res {
		var data [][]any
		if err = json.Unmarshal(b, &data); err != nil {
			return
		}
		if len(data) == 0 {
			err = errors.New("empty data")
			return
		}
		for k, v := range map[string][]any{
			"skills":    {"名前", "記事URL"},
			"character": {"名前", "記事URL", "アイコン画像URL", "レアリティ", "イベント", "ID"},
			"support":   {"名前", "キャラ", "記事URL", "アイコン画像URL", "レアリティ", "タイプ", "イベント", "ID", "シナリオリンクイベント"},
			"choices":   {"名前", "カテゴリ", "キャラ", "発生タイミング", "選択肢", "成功結果", "失敗結果", "ふりがな", "シナリオリンクキャラ", "イベント名"},
		} {
			if reflect.DeepEqual(v, data[0]) {
				switch k {
				case "skills":
					p.skills = data[1:]
				case "character":
					p.character = data[1:]
				case "support":
					p.support = data[1:]
				case "choices":
					choices = data[1:]
				}
			}
		}
	}

	p.data = make(map[string]string)
	for _, i := range p.character {
		if i[1] != "" {
			p.data[i[1].(string)] = i[2].(string)
		}
	}
	for _, i := range p.support {
		if i[1] != "" {
			p.data[i[2].(string)] = i[3].(string)
		}
	}

	if !process {
		return
	}

	for _, i := range choices {
		if i[0] == "" || i[4] == "" || len(strings.Split(i[4].(string), "\n")) == 1 {
			continue
		}

		var event event
		event.Event = i[0].(string)
		event.Character = i[2].(string)
		event.Keyword = i[7].(string)
		switch i[1] {
		case "育成ウマ娘":
			event.Type = "c"
		case "サポートカード":
			event.Type = "s"
		case "メインシナリオ":
			event.Type = "m"
		}

		events = append(events, p.generate(event, i)...)
	}

	for i, event := range events {
		if event.Character == "チーム＜シリウス＞" {
			log.Printf("%s's Rare should be グルSSR, now is %s", event.Character, event.Rare)
			events[i].Rare = "グルSSR"
		}
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].Character == events[j].Character {
			return events[i].Event < events[j].Event
		}
		return events[i].Character < events[j].Character
	})

	return
}

func (p *kamigame) images() error {
	for article, url := range p.data {
		id := regexp.MustCompile(`(\d+).html`).FindStringSubmatch(article)[1]
		if err := downloadImage(url, fmt.Sprintf("public/image/%s.png", id), nil); err != nil {
			return err
		}
	}

	return nil
}

func (p *kamigame) generate(event event, choices []any) (events []event) {
	options, success, failure := split(choices[4]), split(choices[5]), split(choices[6])
	n := len(failure)
	if len(options) != len(success) || (n != 1 && n != len(options)) {
		log.Println("bad data:", choices)
		return
	}

	for i, opt := range options {
		option := option{opt, "", make(map[string]string)}
		if n > i {
			if failure := failure[i]; failure != "" && failure != "-" {
				option.Gain = fmt.Sprintf("成功時：\n%s<hr>失敗時：\n%s", success[i], failure)
			} else {
				option.Gain = success[i]
			}
		} else {
			option.Gain = success[i]
		}
		option.Gain = strings.ReplaceAll(option.Gain, "、", "\n")
		for _, skill := range regexp.MustCompile(`「(.+?)」`).FindAllStringSubmatch(option.Gain, -1) {
			for _, s := range p.skills {
				if str := s[0].(string); str != "" && strings.HasPrefix(skill[1], str) {
					option.Gain = strings.ReplaceAll(option.Gain, skill[0], "『"+skill[1]+"』")
					option.Skill[skill[1]] = fmt.Sprint("https://kamigame.jp", s[1])
					break
				}
			}
		}
		event.Options = append(event.Options, option)
	}

	switch event.Type {
	case "c":
		for _, i := range p.character {
			if event.Character == i[0] {
				event.Article = fmt.Sprint("https://kamigame.jp", i[1])
				event.Image = regexp.MustCompile(`(\d+).html`).FindStringSubmatch(event.Article)[1] + ".png"
				break
			}
		}
		events = append(events, event)
	case "m":
		switch event.Character {
		case "URAファイナルズ":
			event.Image = "ura.png"
		case "アオハル杯":
			event.Image = "aoharu.png"
		case "クライマックス":
			event.Image = "climax.png"
		case "共通":
			event.Image = "rijicho.png"
		}
		events = append(events, event)
	case "s":
		for _, i := range p.support {
			if character := i[1]; character != "" && event.Character == character && strings.Contains(i[6].(string), event.Event) {
				event.Article = fmt.Sprint("https://kamigame.jp", i[2])
				event.Image = regexp.MustCompile(`(\d+).html`).FindStringSubmatch(event.Article)[1] + ".png"
				event.Rare = string([]rune(i[5].(string))[:2]) + i[4].(string)
				events = append(events, event)
			}
		}
	}

	return
}

func split(s any) []string {
	return strings.Split(strings.TrimSpace(s.(string)), "\n")
}
