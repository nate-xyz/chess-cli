package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
	"github.com/notnil/chess"
)

func UpdateGameStatus(s *cv.TextView) {
	var status string

	status += root.currentLocalGame.Status
	root.currentLocalGame.Status = ""

	if root.currentLocalGame.Game.Position().Turn() == chess.White {
		status += "white to move!\n\n"
	} else {
		status += "black to move!\n\n"
	}
	UpdateLegalMoves()
	status += "valid moves: [green]" + strings.Join(root.currentLocalGame.LegalMoves, ", ") + "[white]\n"
	s.SetText(status)
}

func UpdateGameHistory(win *cv.TextView) {
	var text string

	for i := len(root.currentLocalGame.MoveHistoryArray) - 1; i >= 0; i-- {
		text += fmt.Sprintf("%v: %v", i, root.currentLocalGame.MoveHistoryArray[i])
		if i%2 == 0 {
			text += " ⬜\n"
		} else {
			text += " ⬛\n"
		}
	}
	win.SetText(text)
}

func UpdateLegalMoves() {

	root.currentLocalGame.LegalMoves = []string{}

	for _, move := range root.currentLocalGame.Game.ValidMoves() {
		root.currentLocalGame.LegalMoves = append(root.currentLocalGame.LegalMoves, move.String())
	}

}
func LocalGameDoMove() {

	//TODO: handle game end
	//do the move

	err := root.currentLocalGame.Game.MoveStr(root.currentLocalGame.NextMove)

	if err == nil {

		//clear the next move
		root.currentLocalGame.MoveHistoryArray = append(root.currentLocalGame.MoveHistoryArray, root.currentLocalGame.NextMove)
		UpdateGameHistory(root.History)
		UpdateBoard(root.Board, root.currentLocalGame.Game.Position().Turn() == chess.White)
		root.currentLocalGame.NextMove = ""
		UpdateGameStatus(root.Status)
		root.app.GetScreen().Beep()

		//check if game is done

		if root.currentLocalGame.Game.Outcome() != chess.NoOutcome {
			// time.Sleep(2 * time.Second)
			// root.PostStatus.SetText(root.currentLocalGame.Game.Method().String())

			gotoPostLocal()
		}

	} else {
		log.Fatal(err)
	}

}

func UpdateBoard(table *cv.Table, white bool) {
	var FEN string

	if white {
		FEN = root.currentLocalGame.Game.Position().String()
		for i := 0; i < 8; i++ {
			rank := cv.NewTableCell(fmt.Sprintf("%v", i+1))
			file := cv.NewTableCell(string(rune('a' + i)))
			rank.SetAlign(cv.AlignCenter)
			file.SetAlign(cv.AlignCenter)
			table.SetCell(8-i, 0, rank)
			table.SetCell(8-i, 9, rank)
			table.SetCell(0, i+1, file)
			table.SetCell(9, i+1, file)
		}
	} else {
		FEN = root.currentLocalGame.Game.Position().Board().Flip(chess.UpDown).Flip(chess.LeftRight).String()
		for i := 0; i < 8; i++ {
			rank := cv.NewTableCell(fmt.Sprintf("%v", 8-i))
			file := cv.NewTableCell(string(rune('h' - i)))
			rank.SetAlign(cv.AlignCenter)
			file.SetAlign(cv.AlignCenter)
			table.SetCell(8-i, 0, rank)
			table.SetCell(8-i, 9, rank)
			table.SetCell(0, i+1, file)
			table.SetCell(9, i+1, file)
		}
	}

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
		} else if unicode.IsDigit(current_piece) { //full row of nothing
			int_, _ := strconv.Atoi(string(current_piece))
			for i := 1; i <= int_; i++ {
				cell := cv.NewTableCell(" ")
				cell.SetAlign(cv.AlignCenter)
				if square%2 == 0 {
					cell.SetBackgroundColor(tc.NewRGBColor(145, 130, 109))
				} else {
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
			cell.SetAlign(cv.AlignCenter)
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
	root.app.QueueUpdateDraw(func() {})
	// table.SetFixed(8, 8)
}

func UpdateResult(tv *cv.TextView) {
	var status string
	status += root.currentLocalGame.Status
	root.currentLocalGame.Status = ""
	tv.SetText(status)
}
