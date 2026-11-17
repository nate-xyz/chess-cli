package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestGetNews(t *testing.T) {
	now := time.Now()
	ms := func(d time.Duration) int64 { return now.Add(d).UnixMilli() }
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/broadcast":
			// live elite event (two sections: one shown), upcoming, too far off, low tier
			fmt.Fprintf(w, `{"tour":{"id":"live1","name":"Candidates 2026 | Open","tier":5,"dates":[%d,%d],"info":{"format":"8-player double round-robin"}},"round":{"name":"Round 9","ongoing":true}}`+"\n", ms(-48*time.Hour), ms(48*time.Hour))
			fmt.Fprintf(w, `{"tour":{"id":"live2","name":"Candidates 2026 | Women","tier":5,"dates":[%d,%d]}}`+"\n", ms(-48*time.Hour), ms(48*time.Hour))
			fmt.Fprintf(w, `{"tour":{"id":"soon","name":"Sinquefield Cup","tier":5,"dates":[%d,%d]}}`+"\n", ms(10*24*time.Hour), ms(20*24*time.Hour))
			fmt.Fprintf(w, `{"tour":{"id":"far","name":"Next Year Open","tier":5,"dates":[%d]}}`+"\n", ms(200*24*time.Hour))
			fmt.Fprintf(w, `{"tour":{"id":"club","name":"Club Swiss","tier":3,"dates":[%d]}}`+"\n", ms(time.Hour))
		case "/api/broadcast/top":
			fmt.Fprintf(w, `{"past":{"currentPageResults":[{"tour":{"id":"done","name":"Grand Prix Final","tier":5,"dates":[%d,%d]}}]}}`, ms(-10*24*time.Hour), ms(-3*24*time.Hour))
		case "/broadcast/live1/players":
			fmt.Fprint(w, `[{"name":"Caruana, Fabiano","rank":2},{"name":"Sindarov, Javokhir","rank":1}]`)
		case "/broadcast/done/players":
			fmt.Fprint(w, `[{"name":"Firouzja, Alireza","rank":1}]`)
		case "/blog.atom":
			fmt.Fprint(w, `<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom">
<entry><title>New feature</title><published>2026-11-01T10:00:00Z</published></entry>
<entry><title>Older post</title><published>2026-10-01T10:00:00Z</published></entry>
<entry><title>Oldest</title><published>2026-09-01T10:00:00Z</published></entry></feed>`)
		default:
			http.NotFound(w, r)
		}
	}))
	n, err := GetNews()
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Live) != 1 || n.Live[0].Name != "Candidates 2026 | Open" || n.Live[0].Leader != "Sindarov" || !n.Live[0].Live {
		t.Errorf("live: %+v", n.Live)
	}
	if len(n.Upcoming) != 1 || n.Upcoming[0].Name != "Sinquefield Cup" {
		t.Errorf("upcoming: %+v", n.Upcoming)
	}
	if len(n.Finished) != 1 || n.Finished[0].Leader != "Firouzja" {
		t.Errorf("finished: %+v", n.Finished)
	}
	if len(n.Blog) != 2 || n.Blog[0].Title != "New feature" {
		t.Errorf("blog: %+v", n.Blog)
	}
}

func TestPuzzleRequests(t *testing.T) {
	var submitted map[string]any
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/puzzle/daily":
			fmt.Fprint(w, `{"game":{"id":"g1","pgn":"e4 e5"},"puzzle":{"id":"p1","rating":1500,"solution":["g1f3"],"themes":["opening"]}}`)
		case "/api/puzzle/next":
			if r.URL.Query().Get("difficulty") != "harder" {
				http.Error(w, "difficulty", 400)
				return
			}
			fmt.Fprint(w, `{"game":{"id":"g2","pgn":"d4"},"puzzle":{"id":"p2","solution":["d7d5"]}}`)
		case "/api/puzzle/batch/mix":
			if r.Header.Get("Content-Type") != "application/json" || r.Header.Get("Authorization") != "Bearer test-token" {
				http.Error(w, "headers", 400)
				return
			}
			b, _ := io.ReadAll(r.Body)
			json.Unmarshal(b, &submitted)
			fmt.Fprint(w, `{"rounds":[{"id":"p1","win":true,"ratingDiff":9}],"glicko":{"rating":1509.6}}`)
		}
	}))
	p, err := GetDailyPuzzle()
	if err != nil || p.Puzzle.ID != "p1" || p.Game.PGN != "e4 e5" {
		t.Fatalf("daily: %+v %v", p, err)
	}
	if p, err := GetNextPuzzle("harder"); err != nil || p.Puzzle.ID != "p2" {
		t.Fatalf("next: %+v %v", p, err)
	}
	res, err := SubmitPuzzle("p1", true, true)
	if err != nil || res.Rounds[0].RatingDiff != 9 || res.Glicko.Rating < 1509 {
		t.Fatalf("submit: %+v %v", res, err)
	}
	sol := submitted["solutions"].([]any)[0].(map[string]any)
	if sol["id"] != "p1" || sol["win"] != true || sol["rated"] != true {
		t.Errorf("submitted %v", submitted)
	}
}

func TestStreamTv(t *testing.T) {
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tv/blitz/feed" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintln(w, `{"t":"featured","d":{"id":"tv1","orientation":"black","players":[{"color":"white","user":{"name":"A","title":"GM"},"rating":2800,"seconds":180},{"color":"black","user":{"name":"B"},"rating":2700,"seconds":170}],"fen":"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"}}`)
		fmt.Fprintln(w, `{"t":"fen","d":{"fen":"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1","lm":"e2e4","wc":178,"bc":170}}`)
	}))
	var got []WatchUpdate
	if err := StreamTv(context.Background(), "blitz", func(u WatchUpdate) { got = append(got, u) }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Game == nil || got[0].Game.White.Title != "GM" || got[0].Game.Orientation != "black" ||
		got[0].WhiteClock != 180 || got[1].LastMove != "e2e4" || got[1].WhiteClock != 178 {
		t.Errorf("updates: %+v", got)
	}
}

func TestStreamGame(t *testing.T) {
	fakeLichess(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/stream/game/abcd1234":
			fmt.Fprintln(w, `{"id":"abcd1234","speed":"blitz","rated":true,"status":"started","players":{"white":{"user":{"name":"W"},"rating":2000},"black":{"user":{"name":"B"},"rating":2100}}}`)
			fmt.Fprintln(w, `{"fen":"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1","lm":"e2e4","wc":180,"bc":180}`)
			fmt.Fprintln(w, `{"id":"abcd1234","status":"resign","winner":"white","players":{"white":{"user":{"name":"W"}},"black":{"user":{"name":"B"}}}}`)
		case "/game/export/abcd1234":
			fmt.Fprint(w, `{"id":"abcd1234","moves":"e4 e5","clock":{"initial":180,"increment":2},"speed":"blitz"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	var got []WatchUpdate
	if err := StreamGame(context.Background(), "abcd1234", func(u WatchUpdate) { got = append(got, u) }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got[0].Game == nil || got[0].Game.Black.Rating != 2100 || got[1].LastMove != "e2e4" ||
		got[2].Status != "resign" || got[2].Winner != "white" {
		t.Errorf("updates: %+v", got)
	}
	g, err := GetGame("abcd1234")
	if err != nil || g.Moves != "e4 e5" || g.ClockInitial != 180 || g.ClockIncrement != 2 {
		t.Errorf("export: %+v %v", g, err)
	}
	if _, err := GetGame("../account"); err != ErrBadGameID {
		t.Errorf("bad id: %v", err)
	}
	if err := StreamGame(context.Background(), "x/y", func(WatchUpdate) {}); err != ErrBadGameID {
		t.Errorf("bad stream id: %v", err)
	}
	if !strings.Contains(GameIDFrom("https://lichess.org/abcd1234/black"), "abcd1234") {
		t.Error("GameIDFrom")
	}
}
