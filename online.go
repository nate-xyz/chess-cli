package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
	"github.com/notnil/chess"
)

func UpdateLichessTitle() {
	var titlestr string = LichessTitle
	if Online {
		titlestr += " 🟢"
	} else {
		titlestr += " ⚪"

	}
	if UserInfo.ApiToken == "" {
		titlestr += "[red]\nNot logged into lichess.[blue]\nPlease login through your browser.[white]\n"
	} else {
		titlestr += fmt.Sprintf("\n[green]Logged in[white] as: [blue]%s, %s[white]", Username, UserEmail)
	}

	root.LichessTitle.SetText(titlestr)
	root.app.QueueUpdateDraw(func() {}, root.LichessTitle)
}

func UpdateOnline() {
	UpdateLegalMoves()
	UpdateGameHistory(root.OnlineHistory)
	UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)
	//UpdateGameStatus(root.OnlineStatus)
	UpdateOnlineStatus(root.OnlineStatus)
	UpdateUserInfo()
	root.app.QueueUpdateDraw(func() {})
}

func UpdateOnlineStatus(s *cv.TextView) {
	var status string
	var ratestr string

	if BoardFullGame.Rated {
		ratestr = "Rated"
	} else {
		ratestr = "Casual"
	}

	if BoardFullGame.Speed == "correspondence" {
		status += fmt.Sprintf("\n\n%v • %v\n",
			ratestr,
			strings.Title(BoardFullGame.Speed))
	} else {
		status += fmt.Sprintf("\n\n%v+%v • %v • %v\n",
			timeFormat(int64(BoardFullGame.Clock.Initial)),
			BoardFullGame.Clock.Increment/1000,
			ratestr,
			strings.Title(BoardFullGame.Speed))
	}

	if root.currentLocalGame.Game.Position().Turn() == chess.White {
		status += "White's turn. \n\n"
	} else {
		status += "Black's turn. \n\n"
	}

	status += root.currentLocalGame.Status
	root.currentLocalGame.Status = ""

	s.SetText(status)
}

func UpdateUserInfo() {
	var (
		OppName      string
		UserString   string = "\n[blue]%v[white]"
		OppString    string = "\n[red]%v[white]"
		BlackCapture string = strings.Join(root.currentLocalGame.BlackCaptured, "") + " \t"
		WhiteCapture string = strings.Join(root.currentLocalGame.WhiteCaptured, "") + " \t"
	)

	if BoardFullGame.White.Name == Username {
		OppName = BoardFullGame.Black.Name
		UserString = UserString + " (white)\n"
		UserString += WhiteCapture
		OppString = OppString + " (black)\n"
		OppString = BlackCapture + OppString
	} else {
		OppName = BoardFullGame.White.Name
		OppString = OppString + " (white)\n"
		OppString = WhiteCapture + OppString
		UserString = UserString + " (black)\n"
		UserString += BlackCapture
	}

	if CurrentChallenge.Type == 2 {
		OppName = "🤖"
	}

	UserString = fmt.Sprintf(UserString, Username)
	OppString = fmt.Sprintf(OppString, OppName)

	root.OnlineInfoUser.SetText(UserString)
	root.OnlineInfoOppo.SetText(OppString)
}

func UpdateLoaderIcon(i int) int {
	if i > 7 {
		i = 0
	}
	loadingstr := "\n\t ... [red]" + KnightIconMap[i] + "[white] ... \t\n"

	root.LoaderIcon.SetText(loadingstr + loadingstr + loadingstr + loadingstr + loadingstr + loadingstr)

	i++
	if i > 7 {
		i = 0
	}

	root.app.QueueUpdateDraw(func() {})

	return i
}

func UpdateLoaderMsg(msg string) {

	root.LoaderMsg.SetText(msg)
	root.app.QueueUpdateDraw(func() {})

}

func OnlineGameDoMove(move string) error {

	go func() {
		err := root.currentLocalGame.Game.MoveStr(move)
		if err == nil {
			UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)
			root.app.QueueUpdateDraw(func() {}, root.OnlineBoard)
		}
	}()

	err := MakeMove(currentGameID, move) //do the move
	if err != nil {
		return err
	}
	root.app.GetScreen().Beep()

	UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)

	root.currentLocalGame.NextMove = "" //clear the next move
	UpdateGameStatus(root.OnlineStatus)

	return nil
}

func UpdateChessGame() {
	root.currentLocalGame.Game = NewChessGame
}

func UpdateOnlineTimeView() {
	b := int64(BoardFullGame.State.Btime)
	w := int64(BoardFullGame.State.Wtime)
	LiveUpdateOnlineTimeView(b, w)
}

func TimerLoop(d <-chan bool, v *time.Ticker, t *time.Ticker, bi <-chan BothInc) {
	var Btime int64
	var Wtime int64
	var start time.Time
	for {
		select {
		case b := <-bi:
			Wtime = b.wtime
			Btime = b.btime
			start = time.Now()
		case <-d:
			return
		case <-v.C: //every half second
			var currB int64 = Btime
			var currW int64 = Wtime
			if MoveCount >= 2 {
				if MoveCount%2 == 0 {
					currW -= time.Since(start).Milliseconds()
				} else {
					currB -= time.Since(start).Milliseconds()
				}
				LiveUpdateOnlineTimeView(currB, currW)
				root.app.QueueUpdateDraw(func() {}, root.OnlineTimeUser, root.OnlineTimeOppo)
			}
		case <-t.C: //every ms
			var currB int64 = Btime
			var currW int64 = Wtime
			if MoveCount >= 2 {
				if currB < 10000 || currW < 1000 { //start drawing millis when less than ten seconds
					if MoveCount%2 == 0 {
						currW -= time.Since(start).Milliseconds()
						LiveUpdateOnlineTimeView(currB, currW)
					} else {

					}
					LiveUpdateOnlineTimeView(currB, currW)
					root.app.QueueUpdateDraw(func() {}, root.OnlineTimeUser, root.OnlineTimeOppo)
				}
			}

		}
	}
}

func LiveUpdateOnlineTimeView(b int64, w int64) { //MoveCount
	if BoardFullGame.State.Btime == math.MaxInt32 {
		return
	}

	var White bool = BoardFullGame.White.Name == Username
	var UserStr string
	var OppoStr string

	if BoardFullGame.Speed == "correspondence" {
		if White {
			UserStr += (timeFormat(w))
			OppoStr += (timeFormat(b))
		} else {
			UserStr += (timeFormat(b))
			OppoStr += (timeFormat(w))
		}
	} else {
		binc := int64(BoardGameState.Binc)
		winc := int64(BoardGameState.Winc)
		if White {
			UserStr += (timeFormat(w) + fmt.Sprintf("+%d", winc/1000))
			OppoStr += (timeFormat(b) + fmt.Sprintf("+%d", binc/1000))
		} else {
			UserStr += (timeFormat(b) + fmt.Sprintf("+%d", binc/1000))
			OppoStr += (timeFormat(w) + fmt.Sprintf("+%d", winc/1000))
		}
	}

	if MoveCount > 1 {
		if MoveCount%2 == 0 {
			if White {
				UserStr += " ⏲️\t"
				root.OnlineTimeUser.SetBackgroundColor(tc.ColorSeaGreen)
				root.OnlineTimeOppo.SetBackgroundColor(tc.ColorBlack.TrueColor())
			} else {
				OppoStr += " ⏲️\t"
				root.OnlineTimeOppo.SetBackgroundColor(tc.ColorSeaGreen)
				root.OnlineTimeUser.SetBackgroundColor(tc.ColorBlack.TrueColor())
			}
		} else {
			if !White {
				UserStr += " ⏲️\t"
				root.OnlineTimeUser.SetBackgroundColor(tc.ColorSeaGreen)
				root.OnlineTimeOppo.SetBackgroundColor(tc.ColorBlack.TrueColor())
			} else {
				OppoStr += " ⏲️\t"
				root.OnlineTimeOppo.SetBackgroundColor(tc.ColorSeaGreen)
				root.OnlineTimeUser.SetBackgroundColor(tc.ColorBlack.TrueColor())
			}
		}
	}

	root.OnlineTimeUser.SetText(UserStr)
	root.OnlineTimeOppo.SetText(OppoStr)
}

func OnlineTableHandler(row, col int) {
	selectedCell := translateSelectedCell(row, col, BoardFullGame.White.Name == Username)
	if LastSelectedCell.Alg == selectedCell { //toggle selected status of this cell

		root.OnlineBoard.Select(100, 100)
		LastSelectedCell = PiecePosition{-1, -1, "", true, ""}
	} else { //try to do move

		todoMove := LastSelectedCell.Alg + selectedCell
		if contains(root.currentLocalGame.LegalMoves, todoMove) {
			err := OnlineGameDoMove(todoMove)
			if err != nil {
				root.currentLocalGame.Status += fmt.Sprintf("%v", err)
				UpdateGameStatus(root.OnlineStatus)
			}
		}
		//check if select is empty for updateBoard
		symbol := root.OnlineBoard.GetCell(row, col).GetText()
		LastSelectedCell = PiecePosition{row, col, selectedCell, (symbol == EmptyChar), symbol}
	}
	UpdateBoard(root.OnlineBoard, BoardFullGame.White.Name == Username)
}

func UpdateOngoingList() {
	root.OngoingList.Clear()
	GameListIDArr = []string{}
	for i, game := range OngoingGames {
		if contains(GameListIDArr, game.FullID) {
			continue
		}
		GameListIDArr = append(GameListIDArr, game.FullID)
		variant := game.Variant.Name
		opp := game.Opponent.Username
		oppRating := game.Opponent.Rating
		perf := strings.Title(game.Perf)
		listString := fmt.Sprintf("%v vs %v", perf, opp)
		if oppRating > 0 {
			listString += fmt.Sprintf(" (%v)", oppRating)
		}
		if game.SecondsLeft > 0 {
			listString += " " + timeFormat(int64(game.SecondsLeft*1000))
		} else {
			listString += " Unlimited"
		}
		item := cv.NewListItem(listString)
		var text string = variant
		if game.Rated {
			text += " • Rated •"
		} else {
			text += " • Casual •"
		}

		if game.IsMyTurn {
			if game.Color == "black" {
				text += " Black to play, "
			} else {
				text += " White to play, "
			}
			text += fmt.Sprintf("(%v)", Username)
		} else {
			if game.Color == "black" {
				text += "White to play, "
			} else {
				text += "Black to play, "
			}
			text += fmt.Sprintf("(%v)", game.Opponent.ID)
		}
		item.SetSecondaryText(text)
		item.SetShortcut(rune('a' + i))

		root.OngoingList.AddItem(item)
	}
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
	root.app.QueueUpdateDraw(func() {}, table)
}
