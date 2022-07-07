package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

var _ provider = &gamewith{}

type gamewithEvent struct {
	Event     string `json:"e"`
	Character string `json:"n"`
	Type      string `json:"c"`
	Rare      string `json:"l"`
	Article   string `json:"a"`
	Image     string `json:"i"`
	Keyword   string `json:"k"`
	Options   []struct {
		Branch string            `json:"n"`
		Gain   string            `json:"t"`
		Skill  map[string]string `json:"s,omitempty"`
	} `json:"choices"`
}

type gamewithImages struct {
	Support []struct {
		Name  string `json:"n"`
		Rare  string `json:"l"`
		Type  string `json:"c"`
		Image string `json:"i"`
	}
	Character []struct {
		Name  string `json:"n"`
		Rare  string `json:"l"`
		Type  string `json:"c"`
		Image string `json:"i"`
	} `json:"chara"`
}

// https://gamewith.jp/uma-musume/article/show/259587
type gamewith struct {
	data map[string]string
}

func (p gamewith) name() string { return "GameWith" }

func (p *gamewith) events(process bool) (events []event, err error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()

	var id network.RequestID
	done := make(chan struct{})
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			if strings.Contains(ev.Request.URL, "male_event_datas.js") {
				id = ev.RequestID
			}
		case *network.EventLoadingFinished:
			if ev.RequestID == id {
				close(done)
			}
		}
	})

	file, err := os.CreateTemp("", "*.html")
	if err != nil {
		return
	}
	defer os.Remove(file.Name())

	file.WriteString(`
<meta charset="UTF-8">
<script>eventDatas={}</script>
<script src="https://gamewith-tool.s3-ap-northeast-1.amazonaws.com/uma-musume/common_event_datas.js"></script>
<script src="https://gamewith-tool.s3-ap-northeast-1.amazonaws.com/uma-musume/male_event_datas.js"></script>`)
	file.Close()

	if err = chromedp.Run(ctx, chromedp.Navigate(fmt.Sprintf("file:///%s", file.Name()))); err != nil {
		return
	}

	select {
	case <-ctx.Done():
		err = ctx.Err()
		return
	case <-done:
	}

	var imageDatas gamewithImages
	var eventDatas []gamewithEvent
	var linkDatas map[string]string
	if err = chromedp.Run(
		ctx,
		chromedp.Evaluate("imageDatas", &imageDatas),
		chromedp.Evaluate("linkDatas", &linkDatas),
		chromedp.Evaluate("eventDatas['男']", &eventDatas),
	); err != nil {
		return
	}

	p.data = make(map[string]string)
	for _, i := range imageDatas.Support {
		p.data[i.Name+i.Rare] = i.Image
	}
	for _, i := range imageDatas.Character {
		p.data[i.Name] = i.Image
	}

	if !process {
		return
	}

	for _, e := range eventDatas {
		if e.Article != "" {
			e.Article = "https://gamewith.jp/uma-musume/article/show/" + e.Article
		}

		switch e.Type {
		case "c":
			if e.Character == "共通" {
				e.Image = "rijicho.png"
			} else {
				e.Image = p.data[e.Character]
			}
		case "m":
			switch e.Character {
			case "URA":
				e.Image = "ura.png"
			case "アオハル":
				e.Image = "aoharu.png"
			case "クライマックス":
				e.Image = "climax.png"
			}
		case "s":
			e.Image = p.data[e.Character+e.Rare]
		}

		for i, o := range e.Options {
			e.Options[i].Gain = strings.ReplaceAll(o.Gain, "[br]", "\n")
			e.Options[i].Skill = make(map[string]string)
			re := regexp.MustCompile(`『(.+?)』`)
			skills := re.FindAllStringSubmatch(o.Gain, -1)
			for _, skill := range skills {
				if article, ok := linkDatas[skill[1]]; ok {
					e.Options[i].Skill[skill[1]] = "https://gamewith.jp/uma-musume/article/show/" + article
				}
			}
		}

		events = append(events, event(e))
	}

	return
}

func (p *gamewith) images() error {
	for _, i := range p.data {
		if err := downloadImage("https://img.gamewith.jp/article_tools/uma-musume/gacha/"+i, "public/image/"+i, nil); err != nil {
			return err
		}
	}

	return nil
}
