package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

type gamewith struct {
	imageDatas struct {
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
}

func (p *gamewith) fetchData(data bool) error {
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
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	file.WriteString(`
<meta charset="UTF-8">
<script>eventDatas={}</script>
<script src="https://gamewith-tool.s3-ap-northeast-1.amazonaws.com/uma-musume/common_event_datas.js"></script>
<script src="https://gamewith-tool.s3-ap-northeast-1.amazonaws.com/uma-musume/male_event_datas.js"></script>`)
	file.Close()

	if err := chromedp.Run(ctx, chromedp.Navigate(fmt.Sprintf("file:///%s", file.Name()))); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	var eventDatas []gamewithEvent
	var linkDatas map[string]string
	if err := chromedp.Run(
		ctx,
		chromedp.Evaluate("imageDatas", &p.imageDatas),
		chromedp.Evaluate("linkDatas", &linkDatas),
		chromedp.Evaluate("eventDatas['男']", &eventDatas),
	); err != nil {
		return err
	}

	if !data {
		return nil
	}

	var events []event
	for _, e := range eventDatas {
		if e.Article != "" {
			e.Article = "https://gamewith.jp/uma-musume/article/show/" + e.Article
		}

		switch e.Type {
		case "c":
			if e.Character == "共通" {
				e.Image = "rijicho.png"
			} else {
				for _, image := range p.imageDatas.Character {
					if e.Character == image.Name {
						e.Image = image.Image
					}
				}
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
			for _, image := range p.imageDatas.Support {
				if e.Character == image.Name && e.Rare == image.Rare {
					e.Image = image.Image
				}
			}
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

	b, err := json.MarshalIndent(events, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile("uma.json", b, 0777)
}

func (p *gamewith) fetchImage() error {
	if err := os.MkdirAll("public/image", 0777); err != nil {
		log.Fatal(err)
	}

	images := make(map[string]bool)
	for _, i := range p.imageDatas.Support {
		images[i.Image] = true
	}
	for _, i := range p.imageDatas.Character {
		images[i.Image] = true
	}

	for i := range images {
		if err := downloadImage("https://img.gamewith.jp/article_tools/uma-musume/gacha/"+i, "public/image/"+i, nil); err != nil {
			return err
		}
	}

	return nil
}
