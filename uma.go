package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sunshineplan/imgconv"
)

type character string

func parseCharacter(v any) character {
	if s, ok := v.(string); ok {
		c := character(s)
		if scenario, ok := parseScenario(c); ok {
			return character(scenario)
		}
		return c
	}
	return ""
}

func (c *character) UnmarshalText(b []byte) error {
	*c = parseCharacter(string(b))
	return nil
}

type scenario string

const (
	ura       scenario = "URA"
	aoharu    scenario = "アオハル"
	climax    scenario = "クライマックス"
	grandlive scenario = "グランドライブ"
)

var scenarioList = map[scenario]string{
	ura:       "ura.png",
	aoharu:    "aoharu.png",
	climax:    "climax.png",
	grandlive: "grandlive.png",
}

func parseScenario(c character) (scenario, bool) {
	for k := range scenarioList {
		if strings.Contains(string(c), string(k)) {
			return k, true
		}
	}
	return "", false
}

type provider interface {
	name() string
	events(bool) ([]event, error)
	images() error
}

var providers = []provider{&gamewith{}, &gamerch{}, &kamigame{}}

type event struct {
	Event     string    `json:"e"`
	Character character `json:"c"`
	Type      string    `json:"t"`
	Rare      string    `json:"r"`
	Article   string    `json:"a"`
	Image     string    `json:"i"`
	Keyword   string    `json:"k"`
	Options   []struct {
		Branch string            `json:"b"`
		Gain   string            `json:"g"`
		Skill  map[string]string `json:"s,omitempty"`
	} `json:"o"`
}

type option struct {
	Branch string            `json:"b"`
	Gain   string            `json:"g"`
	Skill  map[string]string `json:"s,omitempty"`
}

func exportEvents(events []event, filename string) error {
	b, err := json.MarshalIndent(events, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, b, 0777)
}

var defaultConverter = imgconv.NewOptions().SetResize(72, 0, 0).SetFormat(imgconv.PNG)

func downloadImage(url, path string, converter *imgconv.Options) error {
	if converter == nil {
		converter = defaultConverter
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		log.Println("downloading", url)

		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		img, err := imgconv.Decode(resp.Body)
		if err != nil {
			log.Print(err)
			return nil
		}

		f, err := os.Create(path)
		if err != nil {
			log.Print(err)
			return nil
		}
		defer f.Close()

		if err := converter.Convert(f, img); err != nil {
			log.Print(err)
		}
	} else if err != nil {
		log.Print(err)
	}

	return nil
}
