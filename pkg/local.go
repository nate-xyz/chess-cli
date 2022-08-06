package pkg

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
	"github.com/notnil/chess"
)

func UpdateGameStatus(s *cv.TextView) {
	var status string

	status += Root.currentLocalGame.Status
	Root.currentLocalGame.Status = ""

	if Root.currentLocalGame.Game.Position().Turn() == chess.White {
		status += "White's turn. \n\n"
	} else {
		status += "Black's turn. \n\n"
	}
	UpdateLegalMoves()
	status += "Valid Moves: [green]" + strings.Join(Root.currentLocalGame.LegalMoves, ", ") + "[white]\n"
	s.SetText(status)
}

func UpdateGameHistory(win *cv.TextView) {
	var text string
	moveArr := Root.currentLocalGame.MoveHistoryArray
	pieceArr, err := GetPieceArr(moveArr)
	if err != nil {
		text += fmt.Sprintf("%v", err)
		pieceArr = []string{}
	}
	for i := len(moveArr) - 1; i >= 0; i-- {
		move := moveArr[i]
		text += fmt.Sprintf("\t%v: %v %v", i, pieceArr[i], move)
		if i%2 == 0 {
			text += " ⬜\n"
		} else {
			text += " ⬛\n"
		}
	}
	win.SetText(text)
}

func UpdateLegalMoves() {
	Root.currentLocalGame.LegalMoves = []string{}
	for _, move := range Root.currentLocalGame.Game.ValidMoves() {
		Root.currentLocalGame.LegalMoves = append(Root.currentLocalGame.LegalMoves, move.String())
	}
}

func LocalGameDoMove(m string) error {
	g := Root.currentLocalGame.Game.Clone()
	move, _ := GetMoveType(m, Root.currentLocalGame.Game) //store move
	p := GetPiece(m[2:], g).String()                      //get piece
	err := Root.currentLocalGame.Game.MoveStr(m)          //do the move
	if err == nil {
		if move.HasTag(chess.Check) {
			Root.currentLocalGame.Status += "Check!"
		}
		if move.HasTag(chess.Capture) {
			if len(Root.currentLocalGame.MoveHistoryArray)%2 == 0 { //white
				Root.currentLocalGame.WhiteCaptured = append(Root.currentLocalGame.WhiteCaptured, p)
			} else {
				Root.currentLocalGame.BlackCaptured = append(Root.currentLocalGame.BlackCaptured, p)
			}
		}
		Root.currentLocalGame.MoveHistoryArray = append(Root.currentLocalGame.MoveHistoryArray, m)
		UpdateGameHistory(Root.History)
		UpdateBoard(Root.Board, Root.currentLocalGame.Game.Position().Turn() == chess.White)
		Root.currentLocalGame.NextMove = "" //clear the next move
		UpdateGameStatus(Root.Status)
		UpdateLocalUserInfo()
		UpdateLocalTime()
		Root.App.GetScreen().Beep()
		if Root.currentLocalGame.Game.Outcome() != chess.NoOutcome { //check if game is done
			gotoPostLocal()
		}
		return nil
	}
	return err
}

func UpdateLocalTime() {

}

func UpdateLocalUserInfo() {
	var (
		OppName      string
		UserString   string = "[blue]%v[white]"
		OppString    string = "[red]%v[white]"
		BlackCapture string = strings.Join(Root.currentLocalGame.BlackCaptured, "") + " \t"
		WhiteCapture string = strings.Join(Root.currentLocalGame.WhiteCaptured, "") + " \t"
	)
	var PlayerName string
	if Root.currentLocalGame.Game.Position().Turn() == chess.White {
		OppName = "Black"
		PlayerName = "White"
		UserString = fmt.Sprintf(UserString, PlayerName)
		OppString = fmt.Sprintf(OppString, OppName)
		UserString = fmt.Sprintf("%v\n%v", UserString, WhiteCapture)
		OppString = fmt.Sprintf("%v\n%v", BlackCapture, OppString)
	} else {
		PlayerName = "Black"
		OppName = "White"
		UserString = fmt.Sprintf(UserString, PlayerName)
		OppString = fmt.Sprintf(OppString, OppName)
		UserString = fmt.Sprintf("%v\n%v", UserString, BlackCapture)
		OppString = fmt.Sprintf("%v\n%v", WhiteCapture, OppString)
	}

	Root.InfoUser.SetText(UserString)
	Root.InfoOppo.SetText(OppString)
}

func UpdateBoard(table *cv.Table, white bool) {
	var FEN string
	if white {
		FEN = Root.currentLocalGame.Game.Position().String()
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
		FEN = Root.currentLocalGame.Game.Position().Board().Flip(chess.UpDown).Flip(chess.LeftRight).String()
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

	//check if piece is selected
	if !LastSelectedCell.Empty {
		FEN = Root.currentLocalGame.Game.Position().String()
		fen, err := chess.FEN(FEN)
		if err != nil {
			log.Fatal(err)
			os.Exit(10)
		}
		NewChessGame = chess.NewGame(fen)
		destArr := []string{}
		for _, m := range NewChessGame.ValidMoves() {
			move := m.String()
			if move[0:2] == LastSelectedCell.Alg {
				destArr = append(destArr, move[2:])

			}

		}
		for _, d := range destArr {
			r, c := translateAlgtoCell(d, white)
			cell := table.GetCell(r, c)
			if cell.GetText() == EmptyChar {
				if white {
					cell.SetTextColor(tc.NewRGBColor(255, 248, 220))
				} else {
					cell.SetTextColor(tc.NewRGBColor(18, 18, 18))
				}
				cell.SetText("•")
			} else {
				t := cell.GetText()
				cell.SetAttributes(tc.AttrItalic | tc.AttrBlink)
				cell.SetText(t)
			}

		}
	}

	Root.App.QueueUpdateDraw(func() {}, table)
}

func LocalTableHandler(row, col int) {
	selectedCell := translateSelectedCell(row, col, Root.currentLocalGame.Game.Position().Turn() == chess.White)
	if LastSelectedCell.Alg == selectedCell {
		//toggle selected status of this cell
		Root.Board.Select(100, 100)
		LastSelectedCell = PiecePosition{-1, -1, "", true, ""}
	} else {
		//try to do move
		todoMove := LastSelectedCell.Alg + selectedCell
		if contains(Root.currentLocalGame.LegalMoves, todoMove) {
			err := LocalGameDoMove(todoMove)
			if err != nil {
				Root.currentLocalGame.Status += fmt.Sprintf("%v", err)
				UpdateGameStatus(Root.Status)
			}
		}
		//check if select is empty for updateBoard
		symbol := Root.Board.GetCell(row, col).GetText()
		LastSelectedCell = PiecePosition{row, col, selectedCell, (symbol == EmptyChar), symbol}
	}
	UpdateBoard(Root.Board, Root.currentLocalGame.Game.Position().Turn() == chess.White)
}

func UpdateResult(tv *cv.TextView) {
	var status string
	status += Root.currentLocalGame.Status
	tv.SetText(status)
	Root.currentLocalGame.Status = ""
}
