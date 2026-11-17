package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestStudies(t *testing.T) {
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/study/by/harry":
			fmt.Fprintln(w, `{"id":"old00001","name":"Old","updatedAt":1000}`)
			fmt.Fprintln(w, `{"id":"new00001","name":"New \u001b[31mred","updatedAt":2000}`)
			fmt.Fprintln(w, `{"id":"../evil","name":"Bad id","updatedAt":3000}`)
		case "/api/study/abcd1234.pgn":
			fmt.Fprint(w, "[Event \"x\"]\n\n1. e4 { hi \x1b]0;title\x07 } *")
		default:
			http.NotFound(w, r)
		}
	}))
	list, err := ListStudies("harry")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].ID != "new00001" || list[0].Name != "New [31mred" {
		t.Errorf("studies: %+v", list)
	}
	if _, err := ListStudies("bad/name"); err != ErrBadUsername {
		t.Errorf("bad username: %v", err)
	}
	pgn, err := GetStudyPGN("abcd1234")
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(pgn, "\x1b\x07") || !strings.Contains(pgn, "1. e4") {
		t.Errorf("pgn not cleaned: %q", pgn)
	}
	if _, err := GetStudyPGN("../x"); err != ErrBadStudyID {
		t.Errorf("bad id: %v", err)
	}
}
