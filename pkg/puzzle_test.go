package pkg

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

// Lichess's daily puzzle example: White to move, solution Kxh2 Ng4+ Qxg4 Bxg4 Bxd8.
const dailyPuzzleJSON = `{"game":{"id":"FyCFwzjd","perf":{"key":"blitz","name":"Blitz"},"rated":true,
"players":[{"name":"h21","id":"h21","color":"white","rating":1974},{"name":"ratomir","id":"ratomir","color":"black","rating":1954}],
"pgn":"d4 Nf6 Bg5 d5 Nf3 e6 e3 Nbd7 Bd3 Bd6 O-O O-O Nbd2 c6 Rc1 Re8 c3 e5 dxe5 Nxe5 Nxe5 Bxe5 Bb1 Bxh2+","clock":"3+2"},
"puzzle":{"id":"m5n0s","rating":1957,"plays":85080,"solution":["g1h2","f6g4","d1g4","c8g4","g5d8"],
"themes":["opening","advantage","long","hangingPiece"],"initialPly":23}}`

func TestPuzzleFlow(t *testing.T) {
	var mu sync.Mutex
	var submitted []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/puzzle/daily":
			io.WriteString(w, dailyPuzzleJSON)
		case "/api/puzzle/batch/mix":
			b, _ := io.ReadAll(r.Body)
			mu.Lock()
			submitted = append(submitted, string(b))
			mu.Unlock()
			io.WriteString(w, `{"rounds":[{"id":"m5n0s","win":false,"ratingDiff":-12}],"glicko":{"rating":1488.4}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	oldHost := api.HostUrl
	api.HostUrl = srv.URL
	defer func() { api.HostUrl = oldHost }()

	startTestApp(t, 160, 50)
	api.UserInfo = api.UserConfig{ApiToken: "test"}
	ps := Root.puzzle
	onUI(func() { startPuzzle(true) })
	waitFor(t, "puzzle to load", func() bool { return ps.game != nil })
	onUI(func() {
		if ps.color != chess.White || ps.idx != 0 {
			t.Errorf("color %v idx %d", ps.color, ps.idx)
		}
	})

	// a wrong move is taken back and fails the puzzle (recorded once)
	onUI(func() { ps.tryMove("g1h1") })
	waitFor(t, "failure recorded", func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(submitted) == 1 && strings.Contains(ps.recorded, "-12")
	})
	onUI(func() {
		if !ps.failed || ps.game.Position().Turn() != chess.White {
			t.Error("wrong move not taken back")
		}
	})

	// the right move: the opponent replies by itself
	onUI(func() { ps.tryMove("g1h2") })
	waitFor(t, "opponent reply", func() bool { return ps.idx == 2 })
	onUI(func() { ps.tryMove("d1g4") })
	waitFor(t, "second reply", func() bool { return ps.idx == 4 })
	onUI(func() { ps.tryMove("g5d8") })
	waitFor(t, "done", func() bool { return ps.done })
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(submitted) != 1 || !strings.Contains(submitted[0], `"win":false`) || !strings.Contains(submitted[0], `"rated":true`) {
		t.Errorf("submitted %v", submitted)
	}
}
