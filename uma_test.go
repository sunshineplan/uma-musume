package main

import (
	"log"
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
		"Landscape color": "緑が好きなんだな",           // character
		"常に心にステージを☆":      "勉強も手を抜かないで",         // support
		"乙名史記者の徹底取材":      "現状を真摯に受け止めます",       // URA
		"ついに集まったチームメンバー":  "HOP CHEERS",         // アオハル
		"サプライズ大作戦":        "彼女が素直に意見を聞く人って誰だろう", // クライマックス
		//"新曲プロデュース":        "初心にかえりましょう",         // グランドライブ
	}

	for _, p := range []provider{&gamewith{}, &kamigame{}} {
		events, err := p.events()
		if err != nil {
			t.Errorf("%s: %s", p.name(), err)
			continue
		}
		log.Println(p.name(), len(events))
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
