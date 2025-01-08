package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/sunshineplan/chrome"
)

var _ provider = &gamerch{}

type gamerchEvents struct {
	Cards []struct {
		ID      int `json:"entry_id"`
		Image   string
		Name    character
		Rarity  string
		Support string
		Type    int
	}
	Events []struct {
		ID      int `json:"entry_id"`
		Type    int
		Title   string
		Choices []struct {
			Name    string
			Affects string
		}
	}
	Skills []struct {
		ID   int `json:"entry_id"`
		Name string
	}
}

type gamerchImage struct {
	Name  character
	Rare  string
	Type  string
	Image string
}

// https://gamerch.com/umamusume/event-checker
type gamerch struct {
	data map[int]gamerchImage
}

func (p gamerch) name() string { return "Gamerch" }

func (p *gamerch) events(c *chrome.Chrome) (events []event, err error) {
	if err = c.EnableFetch(func(ev *fetch.EventRequestPaused) bool {
		return ev.ResourceType == network.ResourceTypeDocument ||
			(ev.ResourceType == network.ResourceTypeScript && !regexp.MustCompile("googletag|popin").MatchString(ev.Request.URL)) ||
			ev.ResourceType == network.ResourceTypeXHR
	}); err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(c, time.Minute)
	defer cancel()
	var res gamerchEvents
	done := chrome.ListenEvent(ctx, regexp.MustCompile(`https://cdn\.gamerch\.com/contents/plugin/umamusume/events-\d+\.json`), "GET", true)
	if err = chromedp.Run(ctx, chromedp.Navigate("https://gamerch.com/umamusume/event-checker")); err != nil {
		return
	}
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case event := <-done:
		err = json.Unmarshal(event.Bytes, &res)
	}
	if err != nil {
		return
	}

	p.data = make(map[int]gamerchImage)
	for _, i := range res.Cards {
		image := gamerchImage{i.Name, "", "", i.Image}
		if i.Type == 2 {
			image.Rare = i.Rarity
			image.Type = i.Support
		}
		p.data[i.ID] = image
	}

	for _, e := range res.Events {
		if len(e.Choices) == 1 {
			continue
		}

		var event event
		event.Event = e.Title
		event.Character = p.data[e.ID].Name
		event.Article = fmt.Sprint("https://gamerch.com/umamusume/entry/", e.ID)
		if scenario, ok := parseScenario(event.Character); ok {
			event.Image = scenarioList[scenario]
		} else if event.Character == "あんし～ん笹針師" {
			event.Image = "rijicho.png"
			e.Type = 3
		} else {
			event.Image = fmt.Sprint(e.ID, ".png")
		}

		switch e.Type {
		case 1:
			event.Type = "c"
		case 2:
			if p.data[e.ID].Type == "" {
				continue
			}

			event.Type = "s"
			event.Rare = string([]rune(p.data[e.ID].Type)[:2]) + p.data[e.ID].Rare
		case 3:
			event.Type = "m"
		}

		for _, choice := range e.Choices {
			m := make(map[string]string)
			for _, skill := range regexp.MustCompile(`「(.+?)」`).FindAllStringSubmatch(choice.Affects, -1) {
				for _, s := range res.Skills {
					if strings.HasPrefix(skill[1], s.Name) {
						choice.Affects = strings.ReplaceAll(choice.Affects, skill[0], "『"+skill[1]+"』")
						m[skill[1]] = fmt.Sprint("https://gamerch.com/umamusume/entry/", s.ID)
						break
					}
				}
			}
			event.Options = append(event.Options, option{choice.Name, choice.Affects, m})
		}

		events = append(events, event)
	}

	sort.Slice(events, func(i, j int) bool {
		if events[i].Article == events[j].Article {
			return events[i].Event < events[j].Event
		}
		return events[i].Article < events[j].Article
	})

	return
}

func (p *gamerch) images() error {
	for id, image := range p.data {
		if image.Image == "" {
			continue
		}

		if err := downloadImage(image.Image, fmt.Sprintf("public/image/%d.png", id), nil); err != nil {
			return err
		}
	}

	return nil
}
