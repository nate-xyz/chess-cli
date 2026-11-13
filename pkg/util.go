package pkg

import (
	"fmt"
	"log"
	"math/rand"
	"os"
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

//writes to a local text file
func WriteLocal(title string, payload string) {
	f, err := os.Create(fmt.Sprintf("%s.txt", title))

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(payload)

	if err2 != nil {
		log.Fatal(err2)
	}
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
	Root.App.QueueUpdateDraw(func() {}, board)
}

// func containedInOngoingGames(a []api.OngoingGameInfo, gameid string) bool {
// 	for _, g := range a {
// 		if g.GameID == gameid {
// 			return true
// 		}
// 	}
// 	return false
// }
