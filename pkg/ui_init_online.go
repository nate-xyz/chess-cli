package pkg

import (
	"fmt"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/nate-xyz/chess-cli/api"
)

func (w *WelcomeOnline) Init() *cv.Grid {
	welcomeList := cv.NewList()
	welcomeList.SetHover(true)
	welcomeList.SetWrapAround(true)
	choices := []string{"New Game", "Ongoing Games", "Challenges", "Back", "Quit"}
	explain := []string{
		"Construct a new game request",
		"Select from your active games",
		"Incoming & outgoing challenges",
		"Back to welcome screen",
		"Press to exit"}
	shortcuts := []rune{'n', 'o', 'c', 'b', 'q'}
	selectFunc := []ListSelectedFunc{requireLogin(gotoChallengeConstruction), requireLogin(gotoOngoing), requireLogin(gotoChallenges), gotoWelcome, Root.App.Stop}
	for i := 0; i < len(choices); i++ {
		item := cv.NewListItem(choices[i])
		item.SetSecondaryText(explain[i])
		item.SetShortcut(rune(shortcuts[i]))
		item.SetSelectedFunc(selectFunc[i])
		welcomeList.AddItem(item)
	}

	titleBox := titlePrimitive(LichessTitle, welcomeList)
	welcomeRibbon := ribbonPrimitive(LichessRibbon, welcomeList)
	quoteBox := quoutePrimitive(GetRandomQuote(), welcomeList)

	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	fitMenuColumn(grid, welcomeList, []int{-2, -1}, 1)
	grid.SetRows(-1, -2, -1, 1)
	grid.SetBorders(false)

	grid.AddItem(welcomeList, 1, 1, 1, 1, 0, 0, true)
	grid.AddItem(titleBox, 0, 0, 2, 1, 0, 0, false)
	grid.AddItem(quoteBox, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(welcomeRibbon, 3, 0, 1, 2, 0, 0, false)

	Root.wonline.Title = titleBox
	registerMenus("lichesswelcome", false, welcomeList)

	return grid
}

func (l *Loader) Init() *cv.Grid {
	loaderIconBox := cv.NewTextView()
	loaderIconBox.SetWordWrap(false)
	loaderIconBox.SetDynamicColors(true)
	loaderIconBox.SetTextAlign(cv.AlignCenter)
	loaderIconBox.SetVerticalAlign(cv.AlignMiddle)

	loaderMsgBox := cv.NewTextView()
	loaderMsgBox.SetWordWrap(true)
	loaderMsgBox.SetDynamicColors(true)
	loaderMsgBox.SetTextAlign(cv.AlignCenter)
	loaderMsgBox.SetVerticalAlign(cv.AlignMiddle)

	loaderMsgBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		if event.Key() == tc.KeyEscape || event.Str() == "q" {
			cancelLoader()
			return nil
		}
		return event
	})

	loaderRibbon := ribbonPrimitive(" chess-cli · waiting for Lichess │ Esc to cancel", loaderMsgBox)

	grid := cv.NewGrid()
	grid.SetColumns(-1)
	grid.SetRows(-1, -1, 1)
	grid.SetBorders(false)

	grid.AddItem(loaderMsgBox, 1, 0, 1, 1, 0, 0, true)
	grid.AddItem(loaderIconBox, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(loaderRibbon, 2, 0, 1, 1, 0, 0, false)

	l.Message = loaderMsgBox
	l.Icon = loaderIconBox

	return grid
}

func (g *OnlineGame) Init() cv.Primitive {
	gameBox := NewBoardView(g.OnlineSquareHandler)
	gameBox.EnableSounds()

	chatView := cv.NewTextView()
	chatView.SetDynamicColors(true)
	chatView.SetWordWrap(true)
	chatView.SetScrollable(true)
	chatView.SetBorder(true)
	chatView.SetTitle(" Chat ")
	chatInput := cv.NewInputField()
	chatInput.SetLabel("Say: ")
	chatInput.SetDoneFunc(func(key tc.Key) {
		if key == tc.KeyEnter {
			g.sendChat(chatInput.GetText())
			chatInput.SetText("")
		}
	})

	options := menuList() // filled by UpdateList

	// Tab cycles: board -> chat -> menu -> board
	gameBox.SetTabFunc(func() { Root.App.SetFocus(chatInput) })
	chatInput.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		switch event.Key() {
		case tc.KeyTab:
			Root.App.SetFocus(options)
			return nil
		case tc.KeyEscape:
			Root.App.SetFocus(gameBox)
			return nil
		}
		return event
	})
	options.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		if event.Key() == tc.KeyTab || event.Key() == tc.KeyEscape {
			Root.App.SetFocus(gameBox)
			return nil
		}
		return event
	})

	statusBox := textBox(cv.AlignLeft)
	historyBox := textBox(cv.AlignLeft)
	oppInfoBox := textBox(cv.AlignLeft)
	oppTimerBox := textBox(cv.AlignLeft)
	userTimerBox := textBox(cv.AlignLeft)
	userInfoBox := textBox(cv.AlignLeft)

	side := cv.NewFlex()
	side.SetDirection(cv.FlexRow)
	g.Move = newMoveBox(side, gameBox, func(text string) {
		if uci, err := parseMoveInput(Root.gameState.Game, text); err == nil {
			if err := g.DoMove(uci); err != nil {
				Root.gameState.Status += fmt.Sprintf("%v", err)
				g.UpdateStatus()
			}
		} else {
			Root.gameState.Status += invalidMoveHint(text)
			g.UpdateStatus()
			g.Move.show() // let them try again
		}
	})
	side.AddItem(oppInfoBox, 2, 0, false)
	side.AddItem(oppTimerBox, 0, 0, false)
	side.AddItem(statusBox, 0, 1, false)
	side.AddItem(historyBox, 0, 1, false)
	side.AddItem(chatView, 0, 2, false)
	side.AddItem(chatInput, 1, 0, false)
	side.AddItem(g.Move.input, 0, 0, false)
	side.AddItem(options, 2, 0, false)
	side.AddItem(userTimerBox, 0, 0, false)
	side.AddItem(userInfoBox, 2, 0, false)

	screen := newGameLayout(gameBox, side, gameOnlineRibbonstr)

	g.Side = side
	g.Chat = chatView
	g.ChatInput = chatInput
	g.Board = gameBox
	g.Status = statusBox
	g.History = historyBox
	linkMoveList(historyBox, gameBox)
	g.OppTimer = oppTimerBox
	g.OppInfo = oppInfoBox
	g.UserTimer = userTimerBox
	g.UserInfo = userInfoBox
	g.List = options
	registerMenus("onlinegame", false, options)

	return screen
}

func (pg *OnlinePostGame) Init() cv.Primitive {
	p, screen := newPostGame([]*cv.ListItem{
		menuItem("Rematch", 'r', func() { Root.ongame.rematch() }),
		menuItem("New game", 'n', gotoChallengeConstruction),
		menuItem("Ongoing games", 'o', gotoOngoing),
		menuItem("Lichess menu", 'h', gotoLichessAfterLogin),
		menuItem("Quit", 'q', Root.App.Stop),
	}, postRibbonstr)
	pg.postGame = p
	registerMenus("postonline", false, p.List)
	return screen
}

func (ong *Ongoing) Init() *cv.Grid {
	gameList := cv.NewList()
	gameList.SetHover(false)
	gameList.SetWrapAround(true)
	gameList.SetChangedFunc(func(i int, li *cv.ListItem) {
		gameID := GameListIDArr[i]
		for _, game := range Root.User.OngoingGames {
			if game.FullID == gameID {
				FEN := game.Fen
				var white bool = (game.IsMyTurn && game.Color == "white") || (!game.IsMyTurn && game.Color == "black")
				FENtoBoard(ong.Preview, FEN, white)
			}
		}
	})
	gameList.SetSelectedFunc(func(i int, li *cv.ListItem) {
		gameID := GameListIDArr[i]
		for _, game := range Root.User.OngoingGames {
			if game.FullID == gameID {
				currentGameID = gameID
				startNewOnlineGame()
			}
		}

	})

	options := cv.NewList()
	optionsList := []string{"Back", "Quit"}
	optionsExplain := []string{"Back to the Lichess menu", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoLichessAfterLogin, Root.App.Stop}
	keys := []rune{'b', 'q'}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetShortcut(keys[i])
		item.SetSelectedFunc(optionsFunc[i])
		options.AddItem(item)
	}
	options.SetSelectedFocusOnly(true)
	options.SetHover(true)

	ribbon := ribbonPrimitive(OngoingRibbonstr, gameList)

	title := titlePrimitive("Select an ongoing game.", gameList)

	preview := NewBoardView(nil)
	forwardKeys(preview, gameList)

	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -4)
	grid.SetRows(10, -2, 10, 1)
	grid.SetBorders(false)

	//row col rowSpan colSpan
	grid.AddItem(gameList, 1, 2, 2, 1, 0, 0, true)
	grid.AddItem(title, 0, 0, 1, 3, 0, 0, false)
	grid.AddItem(preview, 1, 0, 1, 2, 0, 0, false)
	grid.AddItem(ribbon, 3, 0, 1, 3, 0, 0, false)
	grid.AddItem(options, 2, 1, 1, 1, 0, 0, false)

	ong.List = gameList
	ong.Preview = preview
	registerMenus("ongoing", true, gameList, options)

	return grid
}

func (c *Challenges) Init() *cv.Grid {
	outgoing := cv.NewList()
	outgoing.AddItem(cv.NewListItem("placeholder out"))
	incoming := cv.NewList()
	incoming.AddItem(cv.NewListItem("placeholder in"))
	incoming.SetSelectedFunc(func(i int, li *cv.ListItem) {
		if i >= len(InChallengeGameID) {
			return
		}
		id := InChallengeGameID[i]
		li.SetSecondaryText("[gray]accepting…[-]")
		go func() {
			err := api.AcceptChallenge(id)
			Root.App.QueueUpdateDraw(func() {
				if err != nil {
					li.SetSecondaryText(fmt.Sprintf("[red]%v[-]", err))
					return
				}
				currentGameID = id
				startNewOnlineGame()
			})
		}()
	})
	// d declines the highlighted incoming challenge
	incoming.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		if event.Str() != "d" {
			return event
		}
		i := incoming.GetCurrentItemIndex()
		if i < 0 || i >= len(InChallengeGameID) {
			return nil
		}
		id := InChallengeGameID[i]
		go func() {
			err := api.DeclineChallenge(id, "")
			Root.App.QueueUpdateDraw(func() {
				if err != nil {
					Root.wonline.UpdateTitle(fmt.Sprintf("\n[red]Decline failed: %v[-]", err))
				}
				gotoChallenges()
			})
		}()
		return nil
	})
	// selecting an outgoing challenge cancels it
	outgoing.SetSelectedFunc(func(i int, li *cv.ListItem) {
		if i >= len(OutChallengeGameID) {
			return
		}
		id := OutChallengeGameID[i]
		li.SetSecondaryText("[gray]cancelling…[-]")
		go func() {
			err := api.CancelChallenge(id)
			Root.App.QueueUpdateDraw(func() {
				if err != nil {
					li.SetSecondaryText(fmt.Sprintf("[red]%v[-]", err))
					return
				}
				gotoChallenges()
			})
		}()
	})

	options := cv.NewList()
	options.SetSelectedFocusOnly(true)
	options.SetHover(true)
	optionsList := []string{"Back", "Quit"}
	optionsExplain := []string{"Back to the Lichess menu", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoLichessAfterLogin, Root.App.Stop}
	keys := []rune{'b', 'q'}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetShortcut(keys[i])
		item.SetSelectedFunc(optionsFunc[i])
		options.AddItem(item)
	}

	outtitle := titlePrimitive("Outgoing challenges\n[gray]Enter cancels[-]", outgoing)
	intitle := titlePrimitive("Incoming challenges\n[gray]Enter accepts, d declines[-]", incoming)
	ribbon := ribbonPrimitive(" chess-cli · challenges │ ↑↓ or 1-9 to pick │ Tab next list │ Ctrl-C quit", incoming)

	grid := cv.NewGrid()
	grid.SetColumns(-1, -3, -3, -1)
	grid.SetRows(-1, -3, 5, 1)
	grid.SetBorders(false)

	grid.AddItem(incoming, 1, 1, 1, 1, 0, 0, true)
	grid.AddItem(outgoing, 1, 2, 1, 1, 0, 0, false)
	grid.AddItem(ribbon, 3, 0, 1, 4, 0, 0, false)
	grid.AddItem(options, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(intitle, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(outtitle, 0, 2, 1, 2, 0, 0, false)

	c.Out = outgoing
	c.In = incoming
	registerMenus("listchallenge", true, incoming, outgoing, options)

	return grid
}
