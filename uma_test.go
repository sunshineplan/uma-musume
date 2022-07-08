package main

import (
	"strings"
	"testing"
)

func searchEvent(event string, events []event) (res []event) {
	for _, i := range events {
		if strings.Contains(i.Event, event) {
			res = append(res, i)
		}
	}
	return
}

func TestUMA(t *testing.T) {
	tc := map[string]string{
		"Landscape color": "緑が好きなんだな",

		"メモリー☆聖地巡礼": "ファールー子！ファールー子！",

		"どこまでも": "全力で走ってみて",

		"常に心にステージを☆": "勉強も手を抜かないで",
	}

	for _, p := range providers {
		events, err := p.events(true)
		if err != nil {
			t.Errorf("%s: %s", p.name(), err)
			continue
		}
		for event, branch := range tc {
			res := searchEvent(event, events)
			if len(res) == 0 {
				t.Errorf("[%s] 0 results of %s", p.name(), event)
				continue
			}
			var found bool
			for _, i := range res[0].Options {
				if strings.Contains(i.Branch, branch) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("[%s] All results not matched %s", p.name(), event)
			}
		}
	}
}
