package pkg

import (
	"fmt"
	"math"
	"strings"
	"time"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func (og *OnlineGame) UpdateAll() {
	Root.gameState.UpdateLegalMoves()
	DrawMoveHistory(og.History)
	DrawBoard(og.Board, og.IsWhite())
	og.UpdateStatus()
	og.UpdateUserInfo()
	og.UpdateList()
}

func (og *OnlineGame) UpdateList() {
	list := og.List
	list.Clear()
	optionsList := []string{"Leave game", "Quit"}
	optionsExplain := []string{"Back to the Lichess menu (the game keeps going)", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoLichessFromGame, Root.App.Stop}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetSelectedFunc(optionsFunc[i])
		list.AddItem(item)
	}

	optionsList = []string{"Takeback", "Abort", "Offer Draw", "Resign", "Claim Draw", "Claim Victory"}
	optionsExplain = []string{"Propose a takeback", "Abort the current game", "Offer a draw to your opponent", "Resign from the current game",
		"Opponent left: end the game as a draw", "Opponent left: claim the win"}
	optionsFunc = []ListSelectedFunc{og.doProposeTakeBack, og.doAbort, og.doOfferDraw, og.doResign, og.doClaimDraw, og.doClaimVictory}
	for i, opt := range optionsList {
		if opt == "Takeback" && Root.gameState.MoveCount == 0 {
			continue
		}
		if opt == "Abort" && Root.gameState.MoveCount >= 2 {
			continue // Lichess only allows aborting before both sides have moved
		}
		if (opt == "Offer Draw" || opt == "Resign") && Root.gameState.MoveCount < 2 {
			continue
		}
		if (opt == "Claim Draw" || opt == "Claim Victory") && !(og.Gone.Gone && og.canClaim) {
			continue
		}
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetSelectedFunc(optionsFunc[i])
		list.InsertItem(1, item)
	}
	og.Side.ResizeItem(list, list.GetItemCount(), 0)
}

func (w *WelcomeOnline) UpdateTitle(msg string) {
	var titlestr string = LichessTitle
	if api.IsOnline() {
		titlestr += " 🟢"
	} else {
		titlestr += " ⚪"
	}
	if Root.User == nil || Root.User.Name == "" {
		titlestr += "[red]\nNot logged in to Lichess.[white]"
		if loginURL != "" {
			titlestr += "\n[blue]Log in through your browser; it should open by itself.[white]\nIf it didn't, open this link:\n" + cv.Escape(loginURL)
		}
	} else {
		titlestr += fmt.Sprintf("\n[green]Logged in[white] as: [blue]%s[white]\n", cv.Escape(Root.User.Name))
	}
	if msg != "" {
		titlestr += msg
	}
	w.Title.SetText(titlestr)
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

	if zenMode() {
		// zen: no game info
	} else if og.Full.Speed == "correspondence" {
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
		status += "White to move.\n"
	} else {
		status += "Black to move.\n"
	}
	if og.Gone.Gone {
		if og.canClaim {
			status += "[yellow]Your opponent left. You can claim victory or a draw from the menu.[-]\n"
		} else {
			status += fmt.Sprintf("[yellow]Your opponent left. You can claim the game in %ds if they don't return.[-]\n", og.Gone.ClaimWinInSeconds)
		}
	}
	og.Status.SetText(status)
}

func (og *OnlineGame) UpdateUserInfo() {
	describe := func(p api.GamePlayer) string {
		if p.AiLevel > 0 {
			return fmt.Sprintf("🤖 Stockfish level %d", p.AiLevel)
		}
		name := p.Name
		if p.Title != "" {
			name = p.Title + " " + name
		}
		name = cv.Escape(name)
		if og.Full.Rated && p.Rating > 0 {
			name += fmt.Sprintf(" (%d", p.Rating)
			if p.Provisional {
				name += "?"
			}
			name += ")"
		}
		return name
	}
	whiteCaptured := strings.Join(Root.gameState.WhiteCaptured, "") + " \t"
	blackCaptured := strings.Join(Root.gameState.BlackCaptured, "") + " \t"
	white := og.IsWhite()
	me, opp := og.Full.White, og.Full.Black
	myColor, oppColor, myCaptured, oppCaptured := "white", "black", whiteCaptured, blackCaptured
	if !white {
		me, opp = opp, me
		myColor, oppColor, myCaptured, oppCaptured = oppColor, myColor, oppCaptured, myCaptured
	}
	meName, oppName := describe(me), describe(opp)
	if zenMode() {
		meName, oppName = "You", "Opponent"
	}
	og.UserInfo.SetText(fmt.Sprintf("\n[blue]%s[white] (%s)\n%s", meName, myColor, myCaptured))
	og.OppInfo.SetText(fmt.Sprintf("%s\n[red]%s[white] (%s)\n", oppCaptured, oppName, oppColor))
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
	return i
}

// DoMove plays a move: it's shown straight away, then sent to Lichess in the
// background. If Lichess rejects it, the board goes back to the last position
// Lichess confirmed. UI goroutine only.
func (og *OnlineGame) DoMove(move string) error {
	if err := Root.gameState.Game.MoveStr(move); err != nil {
		return err
	}
	Root.gameState.NextMove = ""
	DrawBoard(og.Board, og.IsWhite())
	og.UpdateStatus()
	Root.App.GetScreen().Beep()

	gameID := currentGameID
	go func() {
		err := api.MakeMove(gameID, move)
		if err == nil {
			return
		}
		Root.App.QueueUpdateDraw(func() {
			if og.ended || currentGameID != gameID {
				return
			}
			Root.gameState.Status += fmt.Sprintf("[red]Move %s rejected: %v[-]\n", move, err)
			og.applyState(og.State) // back to the last confirmed position
		})
	}()
	return nil
}

func (og *OnlineGame) LiveUpdateTime(b int64, w int64) { //MoveCount
	if og.Full.State.Btime == math.MaxInt32 { // unlimited: no clocks
		if clockBarsOn() {
			og.Board.SetClockBars(clockBar{label: "∞", frac: -1, right: "Unlimited"}, clockBar{label: "∞", frac: -1, right: "Unlimited"})
		} else {
			og.UserTimer.SetText("∞")
			og.OppTimer.SetText("∞ Unlimited")
		}
		return
	}

	var White bool = og.IsWhite()
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
	mine := w
	if !og.IsWhite() {
		mine = b
	}
	myTurn := (Root.gameState.Game.Position().Turn() == chess.White) == og.IsWhite()
	if !og.lowWarned && !og.ended && myTurn && og.Full.Speed != "correspondence" &&
		lowTime(time.Duration(mine)*time.Millisecond, time.Duration(og.Full.Clock.Initial)*time.Millisecond) {
		og.lowWarned = true
		playSound(soundLowTime)
	}
	if clockBarsOn() {
		og.updateClockBars(b, w)
	}
}

// timeControl describes the game's time control, e.g. "10+5".
func (og *OnlineGame) timeControl() string {
	return timeControlText(og.Full.Clock.Initial/1000, og.Full.Clock.Increment/1000, og.Full.Speed, og.Full.DaysPerTurn)
}

// updateClockBars shows the clocks beside the board: you at the bottom.
// b and w are milliseconds.
func (og *OnlineGame) updateClockBars(b, w int64) {
	initial := int64(og.Full.Clock.Initial)
	running := Root.gameState.MoveCount >= 2 && !og.ended
	whiteToMove := Root.gameState.Game.Position().Turn() == chess.White
	bar := func(ms int64, active bool) clockBar {
		c := clockBar{label: clockLabel(int(ms / 1000)), frac: -1, active: active && running}
		if og.Full.Speed != "correspondence" && initial > 0 {
			c.frac = float64(ms) / float64(initial)
		}
		return c
	}
	wc, bc := bar(w, whiteToMove), bar(b, !whiteToMove)
	top, bottom := bc, wc
	if !og.IsWhite() {
		top, bottom = wc, bc
	}
	top.right, bottom.right = og.timeControl(), og.timeControl()
	og.Board.SetClockBars(top, bottom)
}

// OnlineSquareHandler handles a click (or Enter) on a board square in an online game.
// Only your own pieces can be selected, and only on your turn.
func (online *OnlineGame) OnlineSquareHandler(sq chess.Square) {
	white := online.IsWhite()
	myTurn := (Root.gameState.Game.Position().Turn() == chess.White) == white
	if !myTurn {
		online.Board.ClearSelection()
	} else {
		online.Board.HandleSquare(Root.gameState.Game, sq, true, func(uci string) {
			if err := online.DoMove(uci); err != nil {
				Root.gameState.Status += fmt.Sprintf("%v", err)
				online.UpdateStatus()
			}
		})
	}
	DrawBoard(online.Board, white)
}

func (ongoing *Ongoing) UpdateList() {
	ongoing.List.Clear()
	GameListIDArr = []string{}
	for i, game := range Root.User.OngoingGames {
		if contains(GameListIDArr, game.FullID) {
			continue
		}
		GameListIDArr = append(GameListIDArr, game.FullID)
		variant := cv.Escape(game.Variant.Name)
		opp := cv.Escape(game.Opponent.Username)
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
			text += fmt.Sprintf("(%v)", cv.Escape(Root.User.Name))
		} else {
			if game.Color == "black" {
				text += "White to play, "
			} else {
				text += "Black to play, "
			}
			text += fmt.Sprintf("(%v)", cv.Escape(game.Opponent.ID))
		}
		item.SetSecondaryText(text)
		item.SetShortcut(rune('a' + i))

		ongoing.List.AddItem(item)
	}
}

func (c *Challenges) UpdateList() {
	c.In.Clear()
	c.Out.Clear()
	InChallengeGameID = nil
	OutChallengeGameID = nil

	for i, challenge := range Root.User.IncomingChallenges {
		if contains(InChallengeGameID, challenge.ID) {
			continue
		}
		InChallengeGameID = append(InChallengeGameID, challenge.ID)
		variant := cv.Escape(challenge.Variant.Name)
		opp := cv.Escape(challenge.Challenger.Name)
		oppRating := challenge.Challenger.Rating
		perf := cv.Escape(caser.String(challenge.Perf.Name))
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
		text += fmt.Sprintf("%v plays %v", cv.Escape(challenge.Challenger.Name), cv.Escape(challenge.Color))
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
		opp := cv.Escape(challenge.DestUser.Name)
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
