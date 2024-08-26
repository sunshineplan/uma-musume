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

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/sunshineplan/chrome"
)

var _ provider = &kamigame{}

// https://kamigame.jp/umamusume/page/152540608660042049.html
type kamigame struct {
	character, support, skills [][]any
	data                       map[string]string
}

func (p kamigame) name() string { return "kamigame" }

func (p *kamigame) events() (events []event, err error) {
	c := chrome.Headless()
	defer c.Close()
	if err = c.EnableFetch(func(ev *fetch.EventRequestPaused) bool {
		return ev.ResourceType == network.ResourceTypeDocument ||
			(ev.ResourceType == network.ResourceTypeScript && !strings.Contains(ev.Request.URL, "doubleclick")) ||
			ev.ResourceType == network.ResourceTypeXHR
	}); err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(c, time.Minute)
	defer cancel()
	var res [][]byte
	done := chrome.ListenEvent(ctx, "https://kamigame.jp/vls-kamigame-gametool/json", "GET", true)
	go chromedp.Run(ctx, chromedp.Navigate("https://kamigame.jp/umamusume/page/152540608660042049.html"))
	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case event := <-done:
			res = append(res, event.Bytes)
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

	for _, i := range choices {
		if i[0] == "" || i[4] == "" || len(strings.Split(i[4].(string), "\n")) == 1 {
			continue
		}

		var event event
		event.Event = i[0].(string)
		event.Character = parseCharacter(i[2])
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
		if event.Character == "チーム＜シリウス＞" || event.Character == "玉座に集いし者たち" {
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
			if event.Character == parseCharacter(i[0]) {
				event.Article = fmt.Sprint("https://kamigame.jp", i[1])
				if re := regexp.MustCompile(`(\d+).html`); re.MatchString(event.Article) {
					event.Image = re.FindStringSubmatch(event.Article)[1] + ".png"
				} else {
					return
				}
				break
			}
		}
		events = append(events, event)
	case "m":
		if scenario, ok := parseScenario(event.Character); ok {
			event.Image = scenarioList[scenario]
		} else if event.Character == "共通" {
			event.Image = "rijicho.png"
		}
		events = append(events, event)
	case "s":
		for _, i := range p.support {
			if character := parseCharacter(i[1]); character != "" && event.Character == character && strings.Contains(i[6].(string), event.Event) {
				event.Article = fmt.Sprint("https://kamigame.jp", i[2])
				if re := regexp.MustCompile(`(\d+).html`); re.MatchString(event.Article) {
					event.Image = re.FindStringSubmatch(event.Article)[1] + ".png"
				} else {
					return
				}
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
