package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/robertkrimen/otto"
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

var vm = otto.New()

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
			log.Println("Argument:", flag.Arg(0))
			return
		}
	default:
		log.Println("Arguments:", strings.Join(flag.Args(), " "))
		return
	}

	if err := fetchData(data); err != nil {
		log.Fatal(err)
	}

	if image {
		fetchImage()
	}
}

func fetchData(data bool) error {
	src := strings.ReplaceAll(
		gohttp.Get(
			"https://gamewith-tool.s3-ap-northeast-1.amazonaws.com/uma-musume/common_event_datas.js", nil,
		).String(),
		"const",
		"var",
	)
	if _, err := vm.Run(src); err != nil {
		return err
	}

	if err := export("imageDatas", &imageDatas); err != nil {
		return err
	}

	if !data {
		return nil
	}

	if err := export("linkDatas", &linkDatas); err != nil {
		return err
	}

	src = strings.ReplaceAll(
		gohttp.Get(
			"https://gamewith-tool.s3-ap-northeast-1.amazonaws.com/uma-musume/male_event_datas.js", nil,
		).String(),
		"window.eventDatas['男']",
		"var eventDatas",
	)
	if _, err := vm.Run(src); err != nil {
		return err
	}

	if err := export("eventDatas", &eventDatas); err != nil {
		return err
	}

	for _, e := range eventDatas {
		if e.Type == "c" {
			for _, image := range imageDatas.Character {
				if e.Character == image.Name {
					e.Image = image.Image
				}
			}
		} else if e.Type == "s" {
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

	f, err := os.Create("uma.json")
	if err != nil {
		return err
	}
	if _, err := f.Write(b); err != nil {
		return err
	}

	return nil
}

func export(name string, dst interface{}) error {
	value, err := vm.Get(name)
	if err != nil {
		return err
	}

	if value.IsUndefined() {
		return errors.New("undefined value")
	}

	v, err := value.Export()
	if err != nil {
		return err
	}

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, dst)
}

func fetchImage() {
	task := imgconv.New()
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
