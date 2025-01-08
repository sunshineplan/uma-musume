package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sunshineplan/chrome"
)

func main() {
	flag.Parse()

	api := new(gamewith)
	c := chrome.Headless().NoSandbox()
	defer c.Close()
	events, err := api.events(c)
	if err != nil {
		log.Fatal(err)
	}
	b, sum, err := exportEvents(events)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%x", sum)

	switch flag.Arg(0) {
	case "release":
		current, err := os.ReadFile("last")
		if err != nil {
			log.Fatal(err)
		}
		if !bytes.Equal(current, fmt.Appendf(nil, "%x", sum)) {
			log.Fatal("sha256 is not same")
		}
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
