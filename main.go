package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

func main() {
	flag.Parse()

	var json, image bool
	switch flag.NArg() {
	case 0:
		json = true
		image = true
	case 1:
		switch flag.Arg(0) {
		case "json":
			json = true
		case "image":
			image = true
		default:
			log.Fatalln("Argument:", flag.Arg(0))
		}
	default:
		log.Fatalln("Arguments:", strings.Join(flag.Args(), " "))
	}

	var events []event
	var err error
	for _, p := range providers {
		if json {
			events, err = p.events(json)
			if err != nil {
				log.Print(err)
				continue
			}

			if err = exportEvents(events, "uma.json"); err != nil {
				log.Fatal(err)
			}
			os.WriteFile("provider", []byte(p.name()), 0644)
		}
		if image {
			if !json {
				b, _ := os.ReadFile("provider")
				if p.name() != string(b) {
					continue
				} else {
					if _, err = p.events(json); err != nil {
						log.Print(err)
						continue
					}
				}
			}
			if err := os.MkdirAll("public/image", 0644); err != nil {
				log.Fatal(err)
			}

			if err = p.images(); err != nil {
				log.Print(err)
				continue
			}
		}
		break
	}
	if err != nil {
		os.Exit(1)
	}
}
