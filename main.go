package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

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

	var err error
	for _, p := range providers {
		if err = p.fetchData(data); err != nil {
			log.Print(err)
			continue
		}
		if image {
			if err = p.fetchImage(); err != nil {
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
