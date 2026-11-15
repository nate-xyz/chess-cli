package pkg

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/notnil/chess"
)

// TestMain lets the test binary act as a tiny UCI engine (FAKE_UCI=1), so
// the engine code can be tested without Stockfish installed. It plays the
// first legal move in the position it's given.
func TestMain(m *testing.M) {
	if os.Getenv("FAKE_UCI") == "1" {
		fakeEngine()
		return
	}
	soundsMuted.Store(true) // tests never make a noise
	os.Exit(m.Run())
}

func fakeEngine() {
	sc := bufio.NewScanner(os.Stdin)
	game := chess.NewGame()
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 0 {
			continue
		}
		switch f[0] {
		case "uci":
			fmt.Println("id name FakeFish\nuciok")
		case "isready":
			fmt.Println("readyok")
		case "position": // position fen <6 fields> [moves ...]
			fen, _ := chess.FEN(strings.Join(f[2:8], " "))
			game = chess.NewGame(fen)
			if len(f) > 9 {
				for _, uci := range f[9:] {
					m, _ := chess.UCINotation{}.Decode(game.Position(), uci)
					game.Move(m)
				}
			}
		case "go":
			moves := game.ValidMoves()
			if len(moves) == 0 {
				fmt.Println("bestmove (none)")
			} else {
				fmt.Println("bestmove " + chess.UCINotation{}.Encode(game.Position(), moves[0]))
			}
		case "quit":
			return
		}
	}
}

func fakeEnginePath(t *testing.T) string {
	t.Helper()
	t.Setenv("FAKE_UCI", "1")
	return os.Args[0]
}

func TestEngineBestMove(t *testing.T) {
	e, err := StartEngine(fakeEnginePath(t), 3)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	start := chess.StartingPosition().String()
	move, err := e.BestMove(ctx, start, []string{"e2e4"})
	if err != nil {
		t.Fatal(err)
	}
	game := chess.NewGame()
	game.MoveStr("e4")
	if _, err := (chess.UCINotation{}).Decode(game.Position(), move); err != nil {
		t.Errorf("engine move %q isn't legal after 1. e4: %v", move, err)
	}
}

func TestEngineRejectsNonEngine(t *testing.T) {
	if err := checkEngine("/bin/true"); err == nil {
		t.Error("/bin/true accepted as a chess engine")
	}
}

// A game against the (fake) engine on the real UI: the engine replies.
func TestEngineGame(t *testing.T) {
	path := fakeEnginePath(t)
	startTestApp(t, 160, 50)
	onUI(func() { startEngineGame(path, 2, chess.White) })
	g := Root.lgame
	waitFor(t, "engine started", func() bool { return g.engine != nil })
	onUI(func() {
		if err := g.DoMove("e2e4"); err != nil {
			t.Error(err)
		}
	})
	waitFor(t, "engine reply", func() bool { return len(Root.gameState.Game.Moves()) == 2 })
	onUI(func() {
		if g.engineTurn() || !g.whiteAtBottom() {
			t.Error("after the reply it should be white's turn, white at the bottom")
		}
		gotoWelcome() // leaving stops the engine
	})
	waitFor(t, "engine stopped", func() bool { return g.engine == nil })
}
