package pkg

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
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

func translateSelectedCell(row, col int, white bool) string {
	var rank string
	var file string
	if white {
		rank = fmt.Sprintf("%v", 8-(row-1))
		file = string(rune('a' + (col - 1)))
	} else {
		rank = fmt.Sprintf("%v", (row))
		file = string(rune('h' - (col - 1)))
	}
	return file + rank
}

func translateAlgtoCell(alg string, white bool) (r, c int) {
	file := alg[0]
	rank := alg[1]
	var row int
	var col int
	if white {
		row = -int(rank) + 57
		col = int(file) - 96
	} else {
		row = int(rank) - 48
		col = -int(file) + 105
	}
	return row, col
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
	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	for _, move := range moveArr {
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

func FENtoBoard(table *cv.Table, FEN string, white bool) {
	if white {
		for i := 0; i < 8; i++ {
			rank := cv.NewTableCell(fmt.Sprintf("%v", i+1))
			file := cv.NewTableCell(string(rune('a' + i)))
			rank.SetAlign(cv.AlignRight)
			file.SetAlign(cv.AlignRight)
			rank.SetSelectable(false)
			file.SetSelectable(false)
			table.SetCell(8-i, 0, rank)
			table.SetCell(8-i, 9, rank)
			table.SetCell(0, i+1, file)
			table.SetCell(9, i+1, file)
		}
	} else {
		for i := 0; i < 8; i++ {
			rank := cv.NewTableCell(fmt.Sprintf("%v", 8-i))
			file := cv.NewTableCell(string(rune('h' - i)))
			rank.SetAlign(cv.AlignRight)
			file.SetAlign(cv.AlignRight)
			rank.SetSelectable(false)
			file.SetSelectable(false)
			table.SetCell(8-i, 0, rank)
			table.SetCell(8-i, 9, rank)
			table.SetCell(0, i+1, file)
			table.SetCell(9, i+1, file)
		}
	}

	empty := cv.NewTableCell(EmptyChar)
	empty.SetSelectable(false)
	empty.SetTextColor(tc.ColorBlack.TrueColor())
	table.SetCell(0, 0, empty)
	table.SetCell(0, 9, empty)
	table.SetCell(9, 0, empty)
	table.SetCell(9, 9, empty)

	//loop through current FEN and print to board
	square := 0
	col, row := 1, 1
	for _, current_piece := range FEN { //loop to parse the FEN string

		if current_piece == ' ' {
			break
		} else if current_piece == '/' {
			col = 1
			row++
			square++
			continue
		} else if unicode.IsDigit(current_piece) { //nothing
			int_, _ := strconv.Atoi(string(current_piece))
			for i := 1; i <= int_; i++ {
				cell := cv.NewTableCell(EmptyChar)
				cell.SetSelectable(true)
				cell.SetAlign(cv.AlignRight)

				if square%2 == 0 {
					cell.SetTextColor(tc.NewRGBColor(145, 130, 109))
					cell.SetBackgroundColor(tc.NewRGBColor(145, 130, 109))
				} else {
					cell.SetTextColor(tc.NewRGBColor(108, 81, 59))
					cell.SetBackgroundColor(tc.NewRGBColor(108, 81, 59))
				}
				table.SetCell(row, col, cell)
				col++
				square++
			}
			if col > 8 {
				col = 1
			}
			continue
		} else if !unicode.IsDigit(current_piece) {
			cell := cv.NewTableCell(PiecesMap[unicode.ToLower(current_piece)] + " ")
			cell.SetSelectable(true)
			cell.SetAlign(cv.AlignRight)
			if unicode.IsUpper(current_piece) {
				cell.SetTextColor(tc.NewRGBColor(255, 248, 220))
			} else {
				cell.SetTextColor(tc.NewRGBColor(18, 18, 18))
			}
			if square%2 == 0 {
				cell.SetBackgroundColor(tc.NewRGBColor(145, 130, 109))
			} else {
				cell.SetBackgroundColor(tc.NewRGBColor(108, 81, 59))
			}
			table.SetCell(row, col, cell)
			square++
			col++
		} else {
			log.Fatal("error parsing starting FEN")
		}
	}
	Root.App.QueueUpdateDraw(func() {}, table)
}
