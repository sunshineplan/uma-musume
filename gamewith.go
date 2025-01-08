package main

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/sunshineplan/chrome"
)

var _ provider = &gamewith{}

type gamewithEvent struct {
	Event     string    `json:"e"`
	Character character `json:"n"`
	Type      string    `json:"c"`
	Rare      string    `json:"l"`
	Article   string    `json:"a"`
	Image     string    `json:"i"`
	Keyword   string    `json:"k"`
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

func (p *gamewith) events(c *chrome.Chrome) (events []event, err error) {
	ctx, cancel := context.WithTimeout(c, time.Minute)
	defer cancel()
	done := chrome.ListenEvent(ctx, "https://gamewith-tool.s3-ap-northeast-1.amazonaws.com/uma-musume/male_event_datas.js", "GET", false)
	if err = chromedp.Run(ctx, chromedp.Navigate("https://umamusume.sunshineplan.cc/gamewith.html")); err != nil {
		return
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
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

	for _, e := range eventDatas {
		if e.Article != "" {
			e.Article = "https://gamewith.jp/uma-musume/article/show/" + e.Article
		}

		switch e.Type {
		case "c":
			if e.Character == "共通" {
				e.Image = "rijicho.png"
			} else {
				e.Image = p.data[string(e.Character)]
			}
		case "m":
			if scenario, ok := parseScenario(e.Character); ok {
				e.Image = scenarioList[scenario]
			}
		case "s":
			e.Image = p.data[string(e.Character)+e.Rare]
		}

		for i, o := range e.Options {
			e.Options[i].Gain = strings.ReplaceAll(o.Gain, "[br]", "\n")
			e.Options[i].Skill = make(map[string]string)
			for _, skill := range regexp.MustCompile(`『(.+?)』`).FindAllStringSubmatch(o.Gain, -1) {
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
