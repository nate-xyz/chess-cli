package pkg

import (
	"testing"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func TestJoinGames(t *testing.T) {
	// the export knows 1. e4 e5 2. Nf3; we joined live after 1. e4 e5 and saw Nf3 Nc6 Bc4
	earlier, _ := gameFromSAN("", "e4 e5 Nf3")
	joined, _ := gameFromSAN("", "e4 e5")
	fen, _ := chess.FEN(joined.Position().String())
	live := chess.NewGame(fen)
	for _, san := range []string{"Nf3", "Nc6", "Bc4"} {
		if err := live.MoveStr(san); err != nil {
			t.Fatal(err)
		}
	}
	got := joinGames(earlier, live)
	if got == nil || len(got.Moves()) != 5 {
		t.Fatalf("joined game: %v", got)
	}
	if want, _ := gameFromSAN("", "e4 e5 Nf3 Nc6 Bc4"); got.Position().String() != want.Position().String() {
		t.Errorf("got %s", got.Position())
	}
	// an export that doesn't reach anything we saw: nothing to join yet
	behind, _ := gameFromSAN("", "e4")
	if joinGames(behind, live) != nil {
		t.Error("joined games that don't meet")
	}
}

func TestGameIDFrom(t *testing.T) {
	for in, want := range map[string]string{
		"q7ZvsdUF":                          "q7ZvsdUF",
		"https://lichess.org/q7ZvsdUF":      "q7ZvsdUF",
		"https://lichess.org/q7ZvsdUFab12":  "q7ZvsdUF", // player's view
		"lichess.org/q7ZvsdUF/black#12":     "q7ZvsdUF",
		" https://lichess.org/q7ZvsdUF?x=1": "q7ZvsdUF",
	} {
		if got := api.GameIDFrom(in); got != want {
			t.Errorf("GameIDFrom(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCompleteFEN(t *testing.T) {
	// after 1. e4 the pawn on e4 is white's, so black is to move
	if got := completeFEN("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR", "e2e4"); got != "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b - - 0 1" {
		t.Errorf("got %q", got)
	}
	full := "2kr3r/pp3pp1/2p1pnp1/q1npP3/8/bP1P2P1/P1PNQPBP/BK1R3R b - - 0 15"
	if got := completeFEN(full, "e4e5"); got != full {
		t.Errorf("full FEN changed: %q", got)
	}
}
