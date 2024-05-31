package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	flag.Parse()

	api := new(gamewith)
	events, err := api.events()
	if err != nil {
		log.Fatal(err)
	}
	b, sum, err := exportEvents(events)
	if err != nil {
		log.Fatal(err)
	}

	switch flag.Arg(0) {
	case "release":
		if err := os.WriteFile("uma.json", b, 0644); err != nil {
			log.Fatal(err)
		}
	default:
		if err := api.images(); err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile("last", fmt.Appendf(nil, "%x", sum), 0644); err != nil {
			log.Fatal(err)
		}
	}
}
