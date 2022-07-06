package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

var _ provider = &gamerch{}

type gamerch struct {
	data map[int]info
}

type info struct {
	Name  string
	Rare  string
	Type  string
	Image string
}

func (p *gamerch) fetchData(data bool) error {
	var res struct {
		Cards []struct {
			ID      int `json:"entry_id"`
			Image   string
			Name    string
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
			ID    int `json:"entry_id"`
			Image string
			Name  string
		}
	}
	resp, err := http.Get("https://cdn.gamerch.com/contents/plugin/umamusume/events-1656579751.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &res); err != nil {
		return err
	}

	p.data = make(map[int]info)
	for _, i := range res.Cards {
		image := info{i.Name, "", "", i.Image}
		if i.Type == 2 {
			image.Rare = i.Rarity
			image.Type = i.Support
		}
		p.data[i.ID] = image
	}

	if !data {
		return nil
	}

	var events []event
	for _, e := range res.Events {
		if len(e.Choices) == 1 {
			continue
		}

		var event event
		event.Event = e.Title
		event.Character = p.data[e.ID].Name
		event.Article = fmt.Sprint("https://gamerch.com/umamusume/entry/", e.ID)
		switch event.Character {
		case "新設!URAファイナルズ":
			event.Image = "ura.png"
		case "アオハル杯～輝け、チームの絆～":
			event.Image = "aoharu.png"
		case "Make a new track!! ～クライマックス開幕～":
			event.Image = "climax.png"
		case "あんし～ん笹針師":
			event.Image = "rijicho.png"
			e.Type = 3
		default:
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
			re := regexp.MustCompile(`「(.+?)」`)
			skills := re.FindAllStringSubmatch(choice.Affects, -1)
			for _, skill := range skills {
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

	b, err = json.MarshalIndent(events, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile("uma.json", b, 0777)
}

func (p *gamerch) fetchImage() error {
	if err := os.MkdirAll("public/image", 0777); err != nil {
		log.Fatal(err)
	}

	for id, image := range p.data {
		if image.Image == "" {
			continue
		}

		if err := downloadImage(image.Image, fmt.Sprintf("public/image/%d.png", id)); err != nil {
			return err
		}
	}

	return nil
}
