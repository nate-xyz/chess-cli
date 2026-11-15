package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	cv "codeberg.org/tslocum/cview"
	"github.com/nate-xyz/chess-cli/api"
)

// Chess news on the home screen: elite events Lichess is broadcasting (live,
// coming up, just finished, with leaders/winners) and the latest Lichess
// blog posts. Only fetched when logged in to Lichess, so offline use stays
// offline; refreshed at most every 15 minutes, and simply not shown if it
// can't be fetched.

var (
	newsBox     *cv.TextView
	newsFetched time.Time // when the news on screen was fetched
	newsLoading bool
	newsStarted bool // fetched once this run (always done at startup)
)

const newsInterval = 15 * time.Minute

// The last news fetched is saved, so the home screen can show it straight
// away next time while fresh news loads.
type newsCache struct {
	Fetched time.Time
	News    *api.News
}

func newsCachePath() (string, error) {
	dir, err := api.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "news-cache.json"), nil
}

func loadNewsCache() *newsCache {
	path, err := newsCachePath()
	if err != nil {
		return nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var c newsCache
	if json.Unmarshal(b, &c) != nil || c.News == nil {
		return nil
	}
	return &c
}

func saveNewsCache(c newsCache) {
	path, err := newsCachePath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return
	}
	if b, err := json.Marshal(c); err == nil {
		tmp := path + ".tmp"
		if os.WriteFile(tmp, b, 0o600) == nil {
			_ = os.Rename(tmp, path)
		}
	}
}

// showNews puts news on the home screen, noting its age if it's old. It
// leaves the screen alone if nothing changed.
func showNews(n *api.News, fetched time.Time) {
	text := newsText(n)
	if text != "" && time.Since(fetched) > time.Hour {
		text += "\n[#707070]as of " + fetched.Local().Format("Jan 2 15:04") + "[-]"
	}
	if text != newsBox.GetText(false) {
		newsBox.SetText(text)
	}
}

// refreshNews shows the news and fetches fresh news in the background: always
// at startup (the saved news from last run shows meanwhile), then at most
// every 15 minutes. UI goroutine.
func refreshNews() {
	if newsBox == nil || !api.HasValidToken() {
		if newsBox != nil {
			newsBox.SetText("")
		}
		return
	}
	if newsFetched.IsZero() {
		if c := loadNewsCache(); c != nil {
			showNews(c.News, c.Fetched)
			newsFetched = c.Fetched
		}
	}
	if newsLoading || newsStarted && time.Since(newsFetched) < newsInterval {
		return
	}
	newsLoading, newsStarted = true, true
	go func() {
		n, err := api.GetNews()
		if err == nil {
			saveNewsCache(newsCache{Fetched: time.Now(), News: n})
		}
		Root.App.QueueUpdateDraw(func() {
			newsLoading = false
			if err != nil {
				return // keep whatever was shown; try again next time
			}
			newsFetched = time.Now()
			showNews(n, newsFetched)
		})
	}()
}

func newsText(n *api.News) string {
	var b strings.Builder
	line := func(tag, text string) {
		fmt.Fprintf(&b, "%s %s\n", tag, text)
	}
	if len(n.Live)+len(n.Upcoming)+len(n.Finished)+len(n.Blog) == 0 {
		return ""
	}
	b.WriteString("[::b]Chess news[::-]\n\n")
	for _, e := range n.Live {
		text := cv.Escape(eventName(e.Name))
		if e.Round != "" {
			text += " · " + cv.Escape(e.Round)
		}
		if e.Leader != "" {
			text += " · [::b]" + cv.Escape(e.Leader) + "[::-] leads"
		}
		tag := "[#e0a030]● On now[-]"
		if e.Live {
			tag = "[red]● LIVE[-]  "
		}
		line(tag, text)
	}
	for _, e := range n.Upcoming {
		line("[#6fa8dc]◷ Soon[-]  ", cv.Escape(eventName(e.Name))+" · "+dateRange(e.Start, e.End))
	}
	for _, e := range n.Finished {
		text := cv.Escape(eventName(e.Name))
		if e.Leader != "" {
			text = "[::b]" + cv.Escape(e.Leader) + "[::-] wins " + text
		}
		line("[green]✓ Result[-]", text)
	}
	if len(n.Blog) > 0 {
		b.WriteString("\n")
		for _, p := range n.Blog {
			line("[#a0a0a0]Lichess blog[-]", cv.Escape(p.Title)+" [#a0a0a0]· "+p.Date.Format("Jan 2")+"[-]")
		}
	}
	return b.String()
}

// eventName shortens a broadcast name to its event and main section:
// "Total Chess Tour 2026 | Group Stage | Combined" → "Total Chess Tour 2026 · Group Stage".
func eventName(name string) string {
	parts := strings.Split(name, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, " · ")
}

func dateRange(start, end time.Time) string {
	switch {
	case end.IsZero() || start.Format("Jan 2") == end.Format("Jan 2"):
		return start.Format("Jan 2")
	case start.Month() == end.Month():
		return start.Format("Jan 2") + "–" + end.Format("2")
	}
	return start.Format("Jan 2") + " – " + end.Format("Jan 2")
}
