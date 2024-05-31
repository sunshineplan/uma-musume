package main

import (
	"crypto/sha256"
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
	var s string
	switch v := v.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return ""
	}

	c := character(s)
	if scenario, ok := parseScenario(c); ok {
		return character(scenario)
	}
	return c
}

func (c *character) UnmarshalText(b []byte) error {
	*c = parseCharacter(b)
	return nil
}

type scenario string

const (
	ura          scenario = "URA"
	aoharu       scenario = "アオハル"
	climax       scenario = "クライマックス"
	grandlive    scenario = "グランドライブ"
	grandmasters scenario = "グランドマスターズ"
	projectlark  scenario = "プロジェクトL’Arc" // unusual single quote
	uaf          scenario = "UAF"
)

var scenarioList = map[scenario]string{
	ura:          "ura.png",
	aoharu:       "aoharu.png",
	climax:       "climax.png",
	grandlive:    "grandlive.png",
	grandmasters: "grandmasters.png",
	projectlark:  "projectlark.png",
	uaf:          "uaf.png",
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
	events() ([]event, error)
	images() error
}

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

func exportEvents(events []event) (b []byte, sum [32]byte, err error) {
	b, err = json.MarshalIndent(events, "", " ")
	if err != nil {
		return
	}
	sum = sha256.Sum256(b)
	return
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
