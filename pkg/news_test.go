package pkg

import (
	"strings"
	"testing"
	"time"

	"github.com/nate-xyz/chess-cli/api"
)

func TestNewsCacheRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	n := &api.News{Live: []api.Event{{Name: "Candidates 2026 | Open", Leader: "Sindarov", Live: true}},
		Blog: []api.BlogPost{{Title: "Hello", Date: time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)}}}
	saveNewsCache(newsCache{Fetched: time.Now(), News: n})
	c := loadNewsCache()
	if c == nil || len(c.News.Live) != 1 || c.News.Live[0].Leader != "Sindarov" {
		t.Fatalf("cache: %+v", c)
	}
	text := newsText(c.News)
	for _, want := range []string{"LIVE", "Candidates 2026 · Open", "Sindarov", "Hello", "Oct 21"} {
		if !strings.Contains(text, want) {
			t.Errorf("news text missing %q:\n%s", want, text)
		}
	}
}
