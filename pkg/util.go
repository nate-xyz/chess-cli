package pkg

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/notnil/chess"
)

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func GetRandomQuote() string {
	rand.Seed(time.Now().UnixNano())
	minrand := 0
	maxrand := 56
	rand_quote_int := rand.Intn(maxrand-minrand+1) + minrand
	var rand_quote string = RandQuoteMap[rand_quote_int]
	return rand_quote
}

func timeFormat(time int64) string {
	if time == 0 {
		return "0"
	}
	ms := time % 1000
	time /= 1000
	sec := time % 60
	time /= 60
	min := time % 60
	hours := time / 60
	if hours == 0 && min == 0 && sec <= 10 {
		return fmt.Sprintf("%02d:%02d:%03d", min, sec, ms)
	} else if hours == 0 {
		return fmt.Sprintf("%02d:%02d", min, sec)
	}
	days := hours / 24
	hours = hours % 24
	if days == 0 {
		return fmt.Sprintf("%d Hours", hours)
	} else if hours == 0 {
		return fmt.Sprintf("%d Days", days)
	} else {
		return fmt.Sprintf("%d Days %d Hours", days, hours)
	}
}

func GetPiece(p string, g *chess.Game) chess.Piece {
	getFile := map[string]chess.File{
		"a": chess.FileA,
		"b": chess.FileB,
		"c": chess.FileC,
		"d": chess.FileD,
		"e": chess.FileE,
		"f": chess.FileF,
		"g": chess.FileG,
		"h": chess.FileH,
	}
	getRank := map[string]chess.Rank{
		"1": chess.Rank1,
		"2": chess.Rank2,
		"3": chess.Rank3,
		"4": chess.Rank4,
		"5": chess.Rank5,
		"6": chess.Rank6,
		"7": chess.Rank7,
		"8": chess.Rank8,
	}
	board := g.Position().Board()
	file := getFile[string(p[0])]
	rank := getRank[string(p[1])]
	return board.Piece(chess.NewSquare(file, rank))
}

func GetPieceArr(moveArr []string) ([]string, error) {
	pieceArray := []string{}
	if moveArr == nil {
		return pieceArray, nil
	}
	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	for _, move := range moveArr {
		if move == "" {
			continue
		}
		if game.Outcome() == chess.NoOutcome {
			piece := GetPiece(move, game)
			pieceArray = append(pieceArray, piece.String())
			err := game.MoveStr(move)
			if err != nil {
				return pieceArray, err
			}
			continue
		}
	}
	return pieceArray, nil
}

func GetCapturePiecesArr(seq string) error {
	if seq == "" {
		return nil
	}
	Root.gameState.WhiteCaptured = []string{}
	Root.gameState.BlackCaptured = []string{}
	moveArr := strings.Split(seq, " ")
	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	for i, mStr := range moveArr {
		if game.Outcome() == chess.NoOutcome {
			move, _ := GetMoveType(mStr, game)
			if move.HasTag(chess.Capture) {
				//get piece
				p := GetPiece(mStr[2:], game).String()
				if i%2 == 0 {
					Root.gameState.WhiteCaptured = append(Root.gameState.WhiteCaptured, p)
				} else {
					Root.gameState.BlackCaptured = append(Root.gameState.BlackCaptured, p)
				}
			}
			err := game.MoveStr(mStr)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func GetMoveType(movestr string, g *chess.Game) (*chess.Move, error) {
	pos := g.Clone().Position()
	for _, n := range []chess.Notation{chess.LongAlgebraicNotation{}, chess.AlgebraicNotation{}, chess.UCINotation{}} {
		m, err := n.Decode(pos, movestr)
		if err == nil {
			return m, nil
		}
	}
	return new(chess.Move), fmt.Errorf("unable to decode move string")
}

// FENtoBoard shows a position given as FEN on a board (used for previews).
func FENtoBoard(board *BoardView, FEN string, white bool) {
	opt, err := chess.FEN(FEN)
	if err != nil {
		return
	}
	board.SetPosition(chess.NewGame(opt).Position(), white)
	board.SetLastMove(chess.NoSquare, chess.NoSquare)
	board.ClearSelection()
	Root.App.QueueUpdateDraw(func() {}) // full draw, so big pieces are flushed too
}

// func containedInOngoingGames(a []api.OngoingGameInfo, gameid string) bool {
// 	for _, g := range a {
// 		if g.GameID == gameid {
// 			return true
// 		}
// 	}
// 	return false
// }

// parseMoveInput turns a typed move into UCI. It accepts UCI ("e2e4",
// "e7e8q"), UCI with a promotion marker ("e7e8=N", "e7e8N"), standard
// algebraic notation ("Nf3", "exd5", "e8=Q", "O-O", "0-0") and long algebraic
// ("Ng1f3"). Only legal moves are returned.
func parseMoveInput(g *chess.Game, in string) (string, error) {
	in = strings.TrimSpace(in)
	if in == "" {
		return "", fmt.Errorf("empty move")
	}
	pos := g.Position()
	legal := map[string]bool{}
	for _, m := range g.ValidMoves() {
		legal[chess.UCINotation{}.Encode(pos, m)] = true
	}
	try := func(n chess.Notation, s string) (string, bool) {
		m, err := n.Decode(pos, s)
		if err != nil {
			return "", false
		}
		uci := chess.UCINotation{}.Encode(pos, m)
		return uci, legal[uci]
	}
	uciForm := strings.ToLower(strings.NewReplacer("=", "", "-", "", "+", "", "#", "", " ", "").Replace(in))
	if uci, ok := try(chess.UCINotation{}, uciForm); ok {
		return uci, nil
	}
	san := strings.NewReplacer("0-0-0", "O-O-O", "0-0", "O-O").Replace(in)
	for _, n := range []chess.Notation{chess.AlgebraicNotation{}, chess.LongAlgebraicNotation{}} {
		if uci, ok := try(n, san); ok {
			return uci, nil
		}
	}
	return "", fmt.Errorf("not a legal move: %s", in)
}

// invalidMoveHint explains why a typed move was rejected, with a nudge for the
// common case of a promotion typed without its piece.
func invalidMoveHint(move string) string {
	m := strings.ToLower(strings.TrimSpace(move))
	if len(m) == 4 && contains(Root.gameState.LegalMoves, m+"q") {
		return fmt.Sprintf("[yellow]%s is a promotion: add the piece, e.g. %s=Q (Q, R, B or N).[white]\n", move, move)
	}
	return fmt.Sprintf("[red]%q isn't a legal move.[white] Try e2e4, Nf3 or e8=Q.\n", move)
}
