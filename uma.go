package main

import (
	"flag"
	"log"
	"strings"

	"github.com/sunshineplan/utils/executor"
)

type provider interface {
	fetchData(bool) error
	fetchImage() error
}

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

func main() {
	flag.Parse()

	var data, image bool
	switch flag.NArg() {
	case 0:
		data = true
		image = true
	case 1:
		switch flag.Arg(0) {
		case "data":
			data = true
		case "image":
			image = true
		default:
			log.Fatalln("Argument:", flag.Arg(0))
		}
	default:
		log.Fatalln("Arguments:", strings.Join(flag.Args(), " "))
	}

	if _, err := executor.ExecuteSerial(
		[]provider{&gamewith{}},
		func(p provider) (any, error) {
			if err := p.fetchData(data); err != nil {
				return nil, err
			}
			if image {
				return nil, p.fetchImage()
			}
			return nil, nil
		},
	); err != nil {
		log.Fatal(err)
	}

}
