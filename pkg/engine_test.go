package pkg

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/notnil/chess"
)

func TestMain(m *testing.M) {
	soundsMuted.Store(true) // tests never make a noise
	os.Exit(m.Run())
}

// realStockfish is the installed Stockfish; the test is skipped without one.
func realStockfish(t *testing.T) string {
	t.Helper()
	path := findEngine()
	if path == "" {
		t.Skip("Stockfish isn't installed")
	}
	return path
}

func TestEngineBestMove(t *testing.T) {
	e, err := StartEngine(realStockfish(t), 3)
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

// A game against Stockfish on the real UI: the engine replies.
func TestEngineGame(t *testing.T) {
	path := realStockfish(t)
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

// Where installers put Stockfish: the package managers' folders on each
// system, so it's found even before the PATH catches up.
func TestStockfishLocations(t *testing.T) {
	locs := strings.Join(stockfishLocations(), "\n")
	var want []string
	switch runtime.GOOS {
	case "windows":
		want = []string{`WinGet\Links\stockfish.exe`, `scoop\shims\stockfish.exe`, `chocolatey\bin\stockfish.exe`}
	case "darwin":
		want = []string{"/opt/homebrew/bin/stockfish", "/usr/local/bin/stockfish"}
	default:
		want = []string{"/usr/games/stockfish", "/usr/bin/stockfish"}
	}
	for _, w := range want {
		if !strings.Contains(locs, w) {
			t.Errorf("%s not looked in on %s", w, runtime.GOOS)
		}
	}
}
