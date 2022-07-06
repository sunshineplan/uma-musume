package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/sunshineplan/imgconv"
)

type provider interface {
	fetchData(bool) error
	fetchImage() error
}

var providers = []provider{&gamewith{}, &gamerch{}}

type event struct {
	Event     string `json:"e"`
	Character string `json:"c"`
	Type      string `json:"t"`
	Rare      string `json:"r"`
	Article   string `json:"a"`
	Image     string `json:"i"`
	Keyword   string `json:"k"`
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

var image = imgconv.NewOptions()

func init() {
	image.SetResize(72, 0, 0).SetFormat("png")
}

func downloadImage(url, path string) error {
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

		if err := image.Convert(f, img); err != nil {
			log.Print(err)
		}
	} else if err != nil {
		log.Print(err)
	}

	return nil
}
