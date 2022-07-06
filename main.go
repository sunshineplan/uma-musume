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
		events, err = p.events(json)
		if err != nil {
			log.Print(err)
			continue
		}
		if json {
			if err = exportEvents(events); err != nil {
				log.Fatal(err)
			}
		}
		if image {
			if err := os.MkdirAll("public/image", 0777); err != nil {
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
