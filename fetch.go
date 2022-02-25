package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/sunshineplan/gohttp"
	"github.com/sunshineplan/imgconv"
)

type event struct {
	Event     string `json:"e"`
	Character string `json:"n"`
	Type      string `json:"c"`
	Rare      string `json:"l"`
	Article   string `json:"a"`
	Image     string `json:"i"`
	Keyword   string `json:"k"`
	Options   []struct {
		Branch string            `json:"n"`
		Gain   string            `json:"t"`
		Skill  map[string]string `json:"s,omitempty"`
	} `json:"choices"`
}

type realEvent struct {
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

type image struct {
	Name  string `json:"n"`
	Rare  string `json:"l"`
	Type  string `json:"c"`
	Image string `json:"i"`
}

var events []realEvent

var eventDatas []event
var linkDatas map[string]string
var imageDatas struct {
	Support   []image
	Character []image `json:"chara"`
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

	if err := fetchData(data); err != nil {
		log.Fatal(err)
	}

	if image {
		fetchImage()
	}
}

// https://gamewith-tool.s3-ap-northeast-1.amazonaws.com/uma-musume/common_event_datas.js
// https://gamewith-tool.s3-ap-northeast-1.amazonaws.com/uma-musume/male_event_datas.js
func fetchData(data bool) error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, time.Minute)
	defer cancel()

	var id network.RequestID
	done := make(chan struct{})
	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			if strings.Contains(ev.Request.URL, "male_event_datas.js") {
				log.Println("Found:", ev.Request.URL)
				id = ev.RequestID
			}
		case *network.EventLoadingFinished:
			if ev.RequestID == id {
				close(done)
			}
		case *fetch.EventRequestPaused:
			go func() {
				c := chromedp.FromContext(ctx)
				ctx := cdp.WithExecutor(ctx, c.Target)

				if (ev.ResourceType == network.ResourceTypeDocument ||
					ev.ResourceType == network.ResourceTypeScript ||
					ev.ResourceType == network.ResourceTypeXHR) &&
					strings.Contains(ev.Request.URL, "gamewith") &&
					!strings.Contains(ev.Request.URL, "ad/index.") {
					log.Println("Allow:", ev.Request.URL)
					fetch.ContinueRequest(ev.RequestID).Do(ctx)
				} else {
					//log.Println("Block:", ev.Request.URL)
					fetch.FailRequest(ev.RequestID, network.ErrorReasonBlockedByClient).Do(ctx)
				}
			}()
		}
	})

	if err := chromedp.Run(
		ctx,
		fetch.Enable(),
		chromedp.Navigate("https://gamewith.jp/uma-musume/article/show/259587"),
	); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}

	if err := chromedp.Run(
		ctx,
		chromedp.Evaluate("imageDatas", &imageDatas),
		chromedp.Evaluate("linkDatas", &linkDatas),
		chromedp.Evaluate("eventDatas['男']", &eventDatas),
	); err != nil {
		return err
	}

	if !data {
		return nil
	}

	for _, e := range eventDatas {
		switch e.Type {
		case "c":
			if e.Character == "共通" {
				e.Image = "rijicho.png"
			} else {
				for _, image := range imageDatas.Character {
					if e.Character == image.Name {
						e.Image = image.Image
					}
				}
			}
		case "m":
			if e.Character == "クライマックス" {
				e.Image = "climax.png"
			}
		case "s":
			for _, image := range imageDatas.Support {
				if e.Character == image.Name && e.Rare == image.Rare {
					e.Image = image.Image
				}
			}
		}

		for i, o := range e.Options {
			e.Options[i].Gain = strings.ReplaceAll(o.Gain, "[br]", "\n")
			e.Options[i].Skill = make(map[string]string)
			re := regexp.MustCompile(`『(.+?)』`)
			skills := re.FindAllStringSubmatch(o.Gain, -1)
			for _, skill := range skills {
				if article, ok := linkDatas[skill[1]]; ok {
					e.Options[i].Skill[skill[1]] = article
				}
			}
		}

		events = append(events, realEvent(e))
	}

	b, err := json.MarshalIndent(events, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile("uma.json", b, 0777)
}

func fetchImage() {
	task := imgconv.NewOptions()
	task.SetResize(72, 0, 0).SetFormat("png")

	if err := os.MkdirAll("public/image", 0777); err != nil {
		log.Fatal(err)
	}

	images := make(map[string]bool)
	for _, i := range imageDatas.Support {
		images[i.Image] = true
	}
	for _, i := range imageDatas.Character {
		images[i.Image] = true
	}

	url := "https://img.gamewith.jp/article_tools/uma-musume/gacha/"
	for i := range images {
		path := "public/image/" + i
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			log.Println("downloading", url+i)

			resp := gohttp.Get(url+i, nil)
			if resp.Error != nil {
				log.Print(resp.Error)
				continue
			}
			defer resp.Body.Close()

			img, err := imgconv.Decode(resp.Body)
			if err != nil {
				log.Print(err)
				continue
			}

			f, err := os.Create(path)
			if err != nil {
				log.Print(err)
				continue
			}
			defer f.Close()

			if err := task.Convert(f, img); err != nil {
				log.Print(err)
			}
		} else {
			log.Print(err)
		}
	}
}
