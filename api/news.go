package api

import (
	"context"
	"encoding/xml"
	"io"
	"sort"
	"strings"
	"time"
)

// Chess news for the home screen: Lichess's broadcasts of elite events
// (live, coming up, recently finished) and the Lichess blog.

// Event is a broadcast tournament.
type Event struct {
	ID       string
	Name     string
	Start    time.Time
	End      time.Time
	Location string
	Round    string // the current or next round
	Live     bool   // a round is being played right now
	Leader   string // or winner, once finished (individual events only)
	team     bool
}

// BlogPost is a Lichess blog entry.
type BlogPost struct {
	Title string
	Date  time.Time
}

// News is everything for the home screen.
type News struct {
	Live, Upcoming, Finished []Event
	Blog                     []BlogPost
}

type broadcastJSON struct {
	Tour struct {
		ID    string  `json:"id"`
		Name  string  `json:"name"`
		Tier  int     `json:"tier"`
		Dates []int64 `json:"dates"`
		Info  struct {
			Format   string `json:"format"`
			Location string `json:"location"`
		} `json:"info"`
	} `json:"tour"`
	Round struct {
		Name     string `json:"name"`
		Ongoing  bool   `json:"ongoing"`
		StartsAt int64  `json:"startsAt"`
	} `json:"round"`
}

func (b broadcastJSON) event() Event {
	e := Event{ID: b.Tour.ID, Name: b.Tour.Name, Location: b.Tour.Info.Location, Round: b.Round.Name, Live: b.Round.Ongoing}
	if d := b.Tour.Dates; len(d) > 0 {
		e.Start, e.End = time.UnixMilli(d[0]), time.UnixMilli(d[len(d)-1])
	}
	e.team = strings.Contains(strings.ToLower(b.Tour.Info.Format), "team")
	return e
}

// eliteTier is Lichess's top broadcast tier (world-class events).
const eliteTier = 5

// GetNews fetches the elite broadcasts and blog posts. Parts that fail are
// left out; an error is returned only if nothing could be fetched.
func GetNews() (*News, error) {
	n := &News{}
	now := time.Now()
	var firstErr error
	seen := map[string]bool{}

	// official broadcasts: live and upcoming
	official, err := getNDJSON("/api/broadcast?nb=60")
	if err != nil {
		firstErr = err
	}
	for _, b := range official {
		if b.Tour.Tier < eliteTier || seen[b.Tour.ID] {
			continue
		}
		seen[b.Tour.ID] = true
		e := b.event()
		switch {
		case !e.Start.IsZero() && e.Start.After(now):
			if e.Start.Before(now.Add(45 * 24 * time.Hour)) {
				n.Upcoming = append(n.Upcoming, e)
			}
		case e.End.IsZero() || e.End.Add(24*time.Hour).After(now):
			n.Live = append(n.Live, e)
		}
	}
	// recently finished, from the top broadcasts page
	var top struct {
		Past struct {
			Results []broadcastJSON `json:"currentPageResults"`
		} `json:"past"`
	}
	if err := call("GET", "/api/broadcast/top?page=1", nil, &top); err == nil {
		for _, b := range top.Past.Results {
			e := b.event()
			if b.Tour.Tier >= eliteTier && !seen[b.Tour.ID] && e.End.After(now.Add(-21*24*time.Hour)) {
				seen[b.Tour.ID] = true
				n.Finished = append(n.Finished, e)
			}
		}
	} else if firstErr == nil {
		firstErr = err
	}
	n.Live, n.Upcoming, n.Finished = dedupe(n.Live, 3), dedupe(n.Upcoming, 2), dedupe(n.Finished, 2)
	sort.Slice(n.Upcoming, func(i, j int) bool { return n.Upcoming[i].Start.Before(n.Upcoming[j].Start) })
	for _, list := range [][]Event{n.Live, n.Finished} {
		for i := range list {
			if !list[i].team {
				list[i].Leader = leaderOf(list[i].ID)
			}
		}
	}

	if posts, err := getBlog(); err == nil {
		n.Blog = posts
	} else if firstErr == nil {
		firstErr = err
	}
	if len(n.Live)+len(n.Upcoming)+len(n.Finished)+len(n.Blog) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return n, nil
}

// dedupe keeps one entry per event (Lichess lists sections like "| Masters"
// and "| Challengers" separately), preferring the main section, and at most
// max events.
func dedupe(events []Event, max int) []Event {
	var out []Event
	index := map[string]int{}
	for _, e := range events {
		base := strings.TrimSpace(strings.SplitN(e.Name, "|", 2)[0])
		main := strings.Contains(e.Name, "Masters") || strings.Contains(e.Name, "Open") || !strings.Contains(e.Name, "|")
		if i, ok := index[base]; ok {
			if main {
				out[i] = e
			}
			continue
		}
		index[base] = len(out)
		out = append(out, e)
	}
	if len(out) > max {
		out = out[:max]
	}
	return out
}

// leaderOf is the player ranked first in a broadcast, by surname.
func leaderOf(id string) string {
	var players []struct {
		Name  string  `json:"name"`
		Rank  int     `json:"rank"`
		Score float64 `json:"score"`
	}
	if call("GET", "/broadcast/"+id+"/players", nil, &players) != nil {
		return ""
	}
	for _, p := range players {
		if p.Rank == 1 {
			return strings.TrimSpace(strings.SplitN(p.Name, ",", 2)[0])
		}
	}
	return ""
}

func getNDJSON(path string) ([]broadcastJSON, error) {
	req, err := newRequest(context.Background(), "GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/x-ndjson")
	res, err := apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, responseError(res)
	}
	var out []broadcastJSON
	err = eachLine(res.Body, func(line []byte) {
		var b broadcastJSON
		if unmarshal(line, &b) == nil {
			out = append(out, b)
		}
	})
	return out, err
}

// getBlog reads the latest posts from the Lichess blog's Atom feed.
func getBlog() ([]BlogPost, error) {
	req, err := newRequest(context.Background(), "GET", "/blog.atom", nil)
	if err != nil {
		return nil, err
	}
	res, err := apiClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, responseError(res)
	}
	var feed struct {
		Entries []struct {
			Title     string `xml:"title"`
			Published string `xml:"published"`
			Updated   string `xml:"updated"`
		} `xml:"entry"`
	}
	if err := xml.NewDecoder(io.LimitReader(res.Body, 4<<20)).Decode(&feed); err != nil {
		return nil, err
	}
	sanitize(&feed)
	var posts []BlogPost
	for _, e := range feed.Entries {
		date := e.Published
		if date == "" {
			date = e.Updated
		}
		t, _ := time.Parse(time.RFC3339, date)
		posts = append(posts, BlogPost{Title: strings.TrimSpace(e.Title), Date: t})
		if len(posts) == 2 {
			break
		}
	}
	return posts, nil
}
