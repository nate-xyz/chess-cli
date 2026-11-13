package pkg

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

// startTestApp runs the real UI on tcell's mock terminal.
func startTestApp(t *testing.T, w, h int) tc.Screen {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Chdir(t.TempDir())
	s := newMockScreen(t, w, h) // not finalised: the app keeps drawing to it (see below)
	app := cv.NewApplication()
	app.SetScreen(s)
	app.EnableMouse(true)
	Root.App = app
	InitUI()
	// The app and screen are left running until the test binary exits: cview
	// 1.6.3 has unsynchronised shutdown (Run clears its screen field without
	// locking, and Stop finalises the screen while a draw may be running),
	// which the race detector flags. That's in cview, not this code.
	go app.Run()
	waitFor(t, "first draw", func() bool { _, _, rw, _ := Root.nav.GetRect(); return rw == w })
	return s
}

// onUI runs f on the UI goroutine and waits for it.
func onUI(f func()) {
	done := make(chan struct{})
	Root.App.QueueUpdateDraw(func() { f(); close(done) })
	<-done
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		ok := false
		onUI(func() { ok = cond() })
		if ok {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

// clickSquare clicks the centre of a square on a board.
func clickSquare(s tc.Screen, b *BoardView, sq chess.Square) {
	b.mu.Lock()
	r, c := displayOf(sq, b.white)
	x, y := b.ox+c*b.sqW+b.sqW/2, b.oy+r*b.sqH+b.sqH/2
	b.mu.Unlock()
	s.EventQ() <- tc.NewEventMouse(x, y, tc.Button1, 0)
	s.EventQ() <- tc.NewEventMouse(x, y, tc.ButtonNone, 0)
}

// fakeGame is a scripted Lichess game between Harry (white) and bob.
type fakeGame struct {
	mu        sync.Mutex
	moves     []string
	chats     []string
	challenge url.Values
	send      chan string // lines to push down the game stream
}

func (f *fakeGame) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	f.mu.Lock()
	defer f.mu.Unlock()
	switch {
	case r.URL.Path == "/api/board/game/stream/g1":
		f.mu.Unlock()
		fmt.Fprintln(w, `{"type":"gameFull","id":"g1","rated":false,"variant":{"key":"standard","name":"Standard"},`+
			`"clock":{"initial":300000,"increment":3000},"speed":"blitz",`+
			`"white":{"id":"harry","name":"Harry","rating":1500},"black":{"id":"bob","name":"bob","rating":1600},`+
			`"initialFen":"startpos","state":{"type":"gameState","moves":"","wtime":300000,"btime":300000,"winc":3000,"binc":3000,"status":"started"}}`)
		w.(http.Flusher).Flush()
		for {
			select {
			case line := <-f.send:
				fmt.Fprintln(w, line)
				w.(http.Flusher).Flush()
			case <-r.Context().Done():
				f.mu.Lock()
				return
			}
		}
	case strings.HasPrefix(r.URL.Path, "/api/board/game/g1/move/"):
		f.moves = append(f.moves, strings.TrimPrefix(r.URL.Path, "/api/board/game/g1/move/"))
		fmt.Fprint(w, `{"ok":true}`)
	case r.URL.Path == "/api/board/game/g1/chat" && r.Method == "GET":
		fmt.Fprint(w, `[{"user":"bob","text":"have fun [red]"}]`)
	case r.URL.Path == "/api/board/game/g1/chat":
		f.chats = append(f.chats, r.PostForm.Get("text"))
		fmt.Fprint(w, `{"ok":true}`)
	case r.URL.Path == "/api/challenge/bob":
		f.challenge = r.PostForm
		fmt.Fprintln(w, `{"challenge":{"id":"rematch1"}}`)
	default:
		http.NotFound(w, r)
	}
}

func (f *fakeGame) state(moves, extra string) string {
	return fmt.Sprintf(`{"type":"gameState","moves":%q,"wtime":290000,"btime":295000,"winc":3000,"binc":3000,"status":"started"%s}`, moves, extra)
}

func TestOnlineGameEndToEnd(t *testing.T) {
	fake := &fakeGame{send: make(chan string, 10)}
	srv := httptest.NewServer(fake)
	defer srv.Close()
	oldHost := api.HostUrl
	api.HostUrl = srv.URL
	defer func() { api.HostUrl = oldHost }()
	api.UserInfo = api.UserConfig{ApiToken: "test"}

	s := startTestApp(t, 160, 50)
	onUI(func() {
		Root.User = &Login{Name: "Harry"}
		currentGameID = "g1"
		startNewOnlineGame()
	})
	og := Root.ongame
	waitFor(t, "game to load", func() bool { return og.Full.ID == "g1" && og.Board.sqW > 0 })

	// chat history loads, and user text can't inject colour tags
	waitFor(t, "chat history", func() bool { return strings.Contains(og.Chat.GetText(true), "have fun [red]") })

	// play e2e4 with the mouse
	clickSquare(s, og.Board, chess.E2)
	clickSquare(s, og.Board, chess.E4)
	waitFor(t, "move sent", func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return len(fake.moves) == 1 && fake.moves[0] == "e2e4"
	})

	// opponent replies and offers a draw (sent twice: must not stack popups)
	fake.send <- fake.state("e2e4", "")
	fake.send <- fake.state("e2e4 e7e5", `,"bdraw":true`)
	fake.send <- fake.state("e2e4 e7e5", `,"bdraw":true`)
	waitFor(t, "opponent move + draw offer", func() bool {
		return len(Root.gameState.MoveHistoryArray) == 2 && og.popupKind == "draw" && og.PopUp != nil
	})

	// decline it: popup goes away and doesn't come back for the same offer
	onUI(og.doRejectDraw)
	waitFor(t, "popup hidden", func() bool { return og.PopUp == nil })

	// chat
	onUI(func() { og.sendChat("good luck") })
	waitFor(t, "chat sent", func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return len(fake.chats) == 1 && fake.chats[0] == "good luck"
	})
	fake.send <- `{"type":"chatLine","username":"bob","text":"you too","room":"player"}`
	waitFor(t, "chat received", func() bool { return strings.Contains(og.Chat.GetText(true), "you too") })

	// opponent leaves: claim options appear once the countdown passes
	fake.send <- `{"type":"opponentGone","gone":true,"claimWinInSeconds":0}`
	waitFor(t, "claim victory offered", func() bool {
		for i := 0; i < og.List.GetItemCount(); i++ {
			if og.List.GetItem(i).GetMainText() == "Claim Victory" {
				return true
			}
		}
		return false
	})

	// bob resigns: game over screen
	fake.send <- `{"type":"gameState","moves":"e2e4 e7e5","wtime":290000,"btime":295000,"status":"resign","winner":"white"}`
	waitFor(t, "post-game screen", func() bool {
		name, _ := Root.nav.GetFrontPanel()
		return name == "postonline" && og.ended
	})

	// rematch: same clock, colours swapped
	onUI(og.rematch)
	waitFor(t, "rematch challenge", func() bool {
		fake.mu.Lock()
		defer fake.mu.Unlock()
		return fake.challenge != nil
	})
	fake.mu.Lock()
	c := fake.challenge
	fake.mu.Unlock()
	if c.Get("color") != "black" || c.Get("clock.limit") != "300" || c.Get("clock.increment") != "3" || c.Get("rated") != "false" {
		t.Errorf("rematch challenge %v", c)
	}
	onUI(cancelLoader)
}
