package pkg

import (
	"fmt"
	"math"
	"strings"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func (og *OnlineGame) UpdateAll() {
	Root.gameState.UpdateLegalMoves()
	DrawMoveHistory(og.History)
	DrawBoard(og.Board, og.Full.White.Name == Root.User.Name)
	og.UpdateStatus()
	og.UpdateUserInfo()
	og.UpdateList()
	Root.RefreshAll()
}

func (og *OnlineGame) UpdateList() {
	list := og.List
	list.Clear()
	optionsList := []string{"Back", "Quit"}
	optionsExplain := []string{"Go back Home", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoLichessFromGame, Root.App.Stop}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetSelectedFunc(optionsFunc[i])
		list.AddItem(item)
	}

	optionsList = []string{"Takeback", "Abort", "Offer Draw", "Resign"}
	optionsExplain = []string{"Propose a takeback", "Abort the current game", "Offer a draw to your opponent", "Resign from the current game"}
	optionsFunc = []ListSelectedFunc{og.doProposeTakeBack, og.doAbort, og.doOfferDraw, og.doResign}
	for i, opt := range optionsList {
		if opt == "Takeback" && Root.gameState.MoveCount == 0 {
			continue
		}
		if (opt == "Offer Draw" || opt == "Resign") && Root.gameState.MoveCount < 2 {
			continue
		}
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetSelectedFunc(optionsFunc[i])
		list.InsertItem(1, item)
	}

}

func (w *WelcomeOnline) UpdateTitle(msg string) {
	var titlestr string = LichessTitle
	if api.Online {
		titlestr += " 🟢"
	} else {
		titlestr += " ⚪"
	}
	if api.UserInfo.ApiToken == "" {
		titlestr += "[red]\nNot logged into lichess.[blue]\nPlease login through your browser.[white]\nLink should open automatically."
	} else {
		titlestr += fmt.Sprintf("\n[green]Logged in[white] as: [blue]%s, %s[white]\n", Root.User.Name, Root.User.Email)
	}
	if msg != "" {
		titlestr += msg
	}
	w.Title.SetText(titlestr)
	Root.App.QueueUpdateDraw(func() {}, w.Title)
}

func (og *OnlineGame) UpdateStatus() {
	var status string = Root.gameState.Status
	Root.gameState.Status = ""
	var ratestr string

	if og.Full.Rated {
		ratestr = "Rated"
	} else {
		ratestr = "Casual"
	}

	if og.Full.Speed == "correspondence" {
		status += fmt.Sprintf("\n%v • %v %v\n",
			ratestr,
			caser.String(og.Full.Speed),
			currentGameID)
	} else {
		status += fmt.Sprintf("\n%v+%v • %v • %v %v\n",
			timeFormat(int64(og.Full.Clock.Initial)),
			og.Full.Clock.Increment/1000,
			ratestr,
			caser.String(og.Full.Speed),
			currentGameID)
	}

	if Root.gameState.Game.Position().Turn() == chess.White {
		status += "White's turn. \n\n"
	} else {
		status += "Black's turn. \n\n"
	}
	og.Status.SetText(status)
}

func (og *OnlineGame) UpdateUserInfo() {
	var (
		OppName      string = "%s"
		You          string = "%s"
		UserString   string = "\n[blue]%v[white]"
		OppString    string = "\n[red]%v[white]"
		BlackCapture string = strings.Join(Root.gameState.BlackCaptured, "") + " \t"
		WhiteCapture string = strings.Join(Root.gameState.WhiteCaptured, "") + " \t"
	)

	if og.Full.White.Name == Root.User.Name {
		if og.Full.Rated {
			OppName = fmt.Sprintf("%s (%d)", og.Full.Black.Name, og.Full.Black.Rating)
			You = fmt.Sprintf("%s (%d)", Root.User.Name, og.Full.White.Rating)
		} else {
			OppName = og.Full.Black.Name
			You = Root.User.Name
		}
		UserString = UserString + " (white)\n"
		UserString += WhiteCapture
		OppString = OppString + " (black)\n"
		OppString = BlackCapture + OppString
	} else {
		if og.Full.Rated {
			OppName = fmt.Sprintf("%s (%d)", og.Full.White.Name, og.Full.White.Rating)
			You = fmt.Sprintf("%s (%d)", Root.User.Name, og.Full.Black.Rating)
		} else {
			OppName = og.Full.White.Name
			You = Root.User.Name
		}
		OppString = OppString + " (white)\n"
		OppString = WhiteCapture + OppString
		UserString = UserString + " (black)\n"
		UserString += BlackCapture
	}

	if CurrentChallenge.Type == 2 {
		OppName = "🤖"
	}

	UserString = fmt.Sprintf(UserString, You)
	OppString = fmt.Sprintf(OppString, OppName)

	og.UserInfo.SetText(UserString)
	og.OppInfo.SetText(OppString)
}

func (l *Loader) DrawIcon(i int) int {
	if i > 7 {
		i = 0
	}
	loadingstr := "\n\t ... [red]" + KnightIconMap[i] + "[white] ... \t\n"
	l.Icon.SetText(loadingstr + loadingstr + loadingstr + loadingstr + loadingstr + loadingstr)
	i++
	if i > 7 {
		i = 0
	}
	Root.RefreshAll()
	return i
}

func (l *Loader) DrawMessage(msg string) {
	l.Message.SetText(msg)
	Root.RefreshAll()
}

func (og *OnlineGame) DoMove(move string) error {
	go func() {
		err := Root.gameState.Game.MoveStr(move)
		if err == nil {
			DrawBoard(og.Board, og.Full.White.Name == Root.User.Name)
			Root.App.QueueUpdateDraw(func() {}, og.Board)
		}
	}()

	err := api.MakeMove(currentGameID, move) //do the move
	if err != nil {
		return err
	}
	Root.App.GetScreen().Beep()

	DrawBoard(og.Board, og.Full.White.Name == Root.User.Name)

	Root.gameState.NextMove = "" //clear the next move
	og.UpdateStatus()

	return nil
}

func (og *OnlineGame) InitTimeView() {
	b := int64(og.Full.State.Btime)
	w := int64(og.Full.State.Wtime)
	og.LiveUpdateTime(b, w)
}

func (og *OnlineGame) LiveUpdateTime(b int64, w int64) { //MoveCount
	if og.Full.State.Btime == math.MaxInt32 {
		return
	}

	var White bool = og.Full.White.Name == Root.User.Name
	var UserStr string
	var OppoStr string

	if og.Full.Speed == "correspondence" {
		if White {
			UserStr += (timeFormat(w))
			OppoStr += (timeFormat(b))
		} else {
			UserStr += (timeFormat(b))
			OppoStr += (timeFormat(w))
		}
	} else {
		binc := int64(og.State.Binc)
		winc := int64(og.State.Winc)
		if White {
			UserStr += (timeFormat(w) + fmt.Sprintf("+%d", winc/1000))
			OppoStr += (timeFormat(b) + fmt.Sprintf("+%d", binc/1000))
		} else {
			UserStr += (timeFormat(b) + fmt.Sprintf("+%d", binc/1000))
			OppoStr += (timeFormat(w) + fmt.Sprintf("+%d", winc/1000))
		}
	}

	if Root.gameState.MoveCount > 1 {
		if Root.gameState.MoveCount%2 == 0 {
			if White {
				UserStr += " ⏲️\t"
				og.UserTimer.SetBackgroundColor(tc.ColorSeaGreen)
				og.OppTimer.SetBackgroundColor(tc.ColorBlack.TrueColor())
			} else {
				OppoStr += " ⏲️\t"
				og.OppTimer.SetBackgroundColor(tc.ColorSeaGreen)
				og.UserTimer.SetBackgroundColor(tc.ColorBlack.TrueColor())
			}
		} else {
			if !White {
				UserStr += " ⏲️\t"
				og.UserTimer.SetBackgroundColor(tc.ColorSeaGreen)
				og.OppTimer.SetBackgroundColor(tc.ColorBlack.TrueColor())
			} else {
				OppoStr += " ⏲️\t"
				og.OppTimer.SetBackgroundColor(tc.ColorSeaGreen)
				og.UserTimer.SetBackgroundColor(tc.ColorBlack.TrueColor())
			}
		}
	}

	og.UserTimer.SetText(UserStr)
	og.OppTimer.SetText(OppoStr)
}

func (online *OnlineGame) OnlineTableHandler(row, col int) {
	selectedCell := translateSelectedCell(row, col, online.Full.White.Name == Root.User.Name)

	if LastSelectedCell.Alg == selectedCell { //toggle selected status of this cell

		online.Board.Select(100, 100)
		LastSelectedCell = PiecePosition{-1, -1, "", true, ""}
	} else { //try to do move

		todoMove := LastSelectedCell.Alg + selectedCell
		if contains(Root.gameState.LegalMoves, todoMove) {
			err := online.DoMove(todoMove)
			if err != nil {
				Root.gameState.Status += fmt.Sprintf("%v", err)
				online.UpdateStatus()
			}
		}
		//check if select is empty for updateBoard
		symbol := online.Board.GetCell(row, col).GetText()
		LastSelectedCell = PiecePosition{row, col, selectedCell, (symbol == EmptyChar), symbol}
	}
	DrawBoard(online.Board, online.Full.White.Name == Root.User.Name)
}

func (ongoing *Ongoing) UpdateList() {
	ongoing.List.Clear()
	GameListIDArr = []string{}
	for i, game := range Root.User.OngoingGames {
		if contains(GameListIDArr, game.FullID) {
			continue
		}
		GameListIDArr = append(GameListIDArr, game.FullID)
		variant := game.Variant.Name
		opp := game.Opponent.Username
		oppRating := game.Opponent.Rating
		perf := caser.String(game.Perf)
		listString := fmt.Sprintf("%v vs %v", perf, opp)
		if oppRating > 0 {
			listString += fmt.Sprintf(" (%v)", oppRating)
		}
		if game.SecondsLeft > 0 {
			listString += " " + timeFormat(int64(game.SecondsLeft*1000))
		} else {
			listString += " Unlimited"
		}
		listString += fmt.Sprintf(" %v", game.GameID)
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
			text += fmt.Sprintf("(%v)", Root.User.Name)
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

		ongoing.List.AddItem(item)
	}
}

func (c *Challenges) UpdateList() {
	c.In.Clear()
	c.Out.Clear()

	for i, challenge := range Root.User.IncomingChallenges {
		if contains(InChallengeGameID, challenge.ID) {
			continue
		}
		InChallengeGameID = append(InChallengeGameID, challenge.ID)
		variant := challenge.Variant.Name
		opp := challenge.Challenger.Name
		oppRating := challenge.Challenger.Rating
		perf := caser.String(challenge.Perf.Name)
		listString := fmt.Sprintf("%v challenge from %v", perf, opp)
		if oppRating > 0 {
			listString += fmt.Sprintf(" (%v) ", oppRating)
		}
		listString += challenge.Speed
		item := cv.NewListItem(listString)
		var text string = variant
		if challenge.Rated {
			text += " • Rated • "
		} else {
			text += " • Casual • "
		}
		text += fmt.Sprintf("%v plays %v", challenge.Challenger.Name, challenge.Color)
		item.SetSecondaryText(text)
		item.SetShortcut(rune('a' + i))
		c.In.AddItem(item)
	}
	for i, challenge := range Root.User.OutgoingChallenges {
		if contains(OutChallengeGameID, challenge.ID) {
			continue
		}
		OutChallengeGameID = append(OutChallengeGameID, challenge.ID)
		variant := challenge.Variant.Name
		opp := challenge.DestUser.Name
		oppRating := challenge.DestUser.Rating
		perf := caser.String(challenge.Perf.Name)
		listString := fmt.Sprintf("%v challenge to %v", perf, opp)
		if oppRating > 0 {
			listString += fmt.Sprintf(" (%v)", oppRating)
		}
		listString += challenge.Speed
		item := cv.NewListItem(listString)
		var text string = variant
		if challenge.Rated {
			text += " • Rated • "
		} else {
			text += " • Casual • "
		}
		text += fmt.Sprintf("%v plays %v", challenge.Challenger.Name, challenge.Color)
		item.SetSecondaryText(text)
		item.SetShortcut(rune('a' + i))
		c.Out.AddItem(item)
	}

}

func (pg *OnlinePostGame) UpdateResult() {
	pg.Result.SetText(Root.gameState.Status)
	Root.gameState.Status = ""
}
