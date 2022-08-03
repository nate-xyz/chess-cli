package main

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

	status += root.currentLocalGame.Status
	root.currentLocalGame.Status = ""

	if root.currentLocalGame.Game.Position().Turn() == chess.White {
		status += "White's turn. \n\n"
	} else {
		status += "Black's turn. \n\n"
	}
	UpdateLegalMoves()
	status += "Valid Moves: [green]" + strings.Join(root.currentLocalGame.LegalMoves, ", ") + "[white]\n"
	s.SetText(status)
}

func UpdateGameHistory(win *cv.TextView) {
	var text string
	moveArr := root.currentLocalGame.MoveHistoryArray
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
	root.currentLocalGame.LegalMoves = []string{}
	for _, move := range root.currentLocalGame.Game.ValidMoves() {
		root.currentLocalGame.LegalMoves = append(root.currentLocalGame.LegalMoves, move.String())
	}
}

func LocalGameDoMove(m string) error {
	//do the move
	move, _ := GetMoveType(m, root.currentLocalGame.Game)
	err := root.currentLocalGame.Game.MoveStr(m)
	if err == nil {

		if move.HasTag(chess.Check) {
			root.currentLocalGame.Status += "Check!"
		}

		if move.HasTag(chess.Capture) {
			//get piece
			p := GetPiece(m[2:], root.currentLocalGame.Game.Clone()).String()
			if len(root.currentLocalGame.MoveHistoryArray)%2 == 0 {
				root.currentLocalGame.WhiteCaptured = append(root.currentLocalGame.WhiteCaptured, p)
			} else {
				root.currentLocalGame.BlackCaptured = append(root.currentLocalGame.BlackCaptured, p)
			}
		}

		root.currentLocalGame.MoveHistoryArray = append(root.currentLocalGame.MoveHistoryArray, m)
		UpdateGameHistory(root.History)
		UpdateBoard(root.Board, root.currentLocalGame.Game.Position().Turn() == chess.White)
		root.currentLocalGame.NextMove = "" //clear the next move
		UpdateGameStatus(root.Status)
		UpdateLocalTime()
		root.app.GetScreen().Beep()
		if root.currentLocalGame.Game.Outcome() != chess.NoOutcome { //check if game is done
			gotoPostLocal()
		}
		return nil
	}
	return err
}

func UpdateLocalTime() {
	t := root.Time
	var txt string
	txt += "\nBLACK CAPTURE:\n"
	for _, p := range root.currentLocalGame.BlackCaptured {
		txt += p
	}
	txt += "\t"
	txt += "\nWHITE CAPTURE:\n"
	for _, wp := range root.currentLocalGame.WhiteCaptured {
		txt += wp
	}
	txt += "\t"
	t.SetText(txt)
}

func UpdateBoard(table *cv.Table, white bool) {
	var FEN string
	if white {
		FEN = root.currentLocalGame.Game.Position().String()
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
		FEN = root.currentLocalGame.Game.Position().Board().Flip(chess.UpDown).Flip(chess.LeftRight).String()
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
		FEN = root.currentLocalGame.Game.Position().String()
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
		//var text string

		for _, d := range destArr {
			r, c := translateAlgtoCell(d, white)
			//text += fmt.Sprintf("%v: %v, %v\n", d, r, c)

			cell := table.GetCell(r, c)
			if cell.GetText() == EmptyChar {
				if white {
					cell.SetTextColor(tc.NewRGBColor(255, 248, 220))
				} else {
					cell.SetTextColor(tc.NewRGBColor(18, 18, 18))
				}
				cell.SetText("•")
				//text += "set text\n"
			} else {
				t := cell.GetText()
				//text += t
				cell.SetAttributes(tc.AttrItalic | tc.AttrBlink)
				cell.SetText(t)
			}

		}

		//root.Time.SetText(text)
	}

	root.app.QueueUpdateDraw(func() {}, table)
}

func LocalTableHandler(row, col int) {
	//SelectionCount++
	//var text string
	// if row == 0 && (col == 0 || col == 9) {
	// 	root.Board.Select(100, 100)
	// }

	// if row == 9 && (col == 0 || col == 9) {
	// 	root.Board.Select(100, 100)
	// }

	selectedCell := translateSelectedCell(row, col, root.currentLocalGame.Game.Position().Turn() == chess.White)
	if LastSelectedCell.Alg == selectedCell {
		//toggle selected status of this cell
		root.Board.Select(100, 100)
		LastSelectedCell = PiecePosition{-1, -1, "", true, ""}
	} else {
		//try to do move
		todoMove := LastSelectedCell.Alg + selectedCell
		if contains(root.currentLocalGame.LegalMoves, todoMove) {
			err := LocalGameDoMove(todoMove)
			if err != nil {
				root.currentLocalGame.Status += fmt.Sprintf("%v", err)
				UpdateGameStatus(root.Status)
			}
		}
		//check if select is empty for updateBoard
		symbol := root.Board.GetCell(row, col).GetText()
		LastSelectedCell = PiecePosition{row, col, selectedCell, (symbol == EmptyChar), symbol}
	}
	//text += selectedCell
	//text += fmt.Sprintf("\n%d %d, %d", row, col, SelectionCount)
	//root.Time.SetText(text)
	UpdateBoard(root.Board, root.currentLocalGame.Game.Position().Turn() == chess.White)
	//root.app.QueueUpdateDraw(func() {}, root.Time)
}

func UpdateResult(tv *cv.TextView) {
	var status string
	status += root.currentLocalGame.Status
	root.currentLocalGame.Status = ""
	tv.SetText(status)
}
