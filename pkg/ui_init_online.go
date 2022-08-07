package pkg

import (
	"fmt"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
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
	selectFunc := []ListSelectedFunc{gotoChallengeConstruction, gotoOngoing, gotoChallenges, gotoWelcome, Root.App.Stop}
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
	grid.SetRows(-1, -2, -1, 1)
	grid.SetBorders(false)

	grid.AddItem(welcomeList, 1, 1, 1, 1, 0, 0, true)
	grid.AddItem(titleBox, 0, 0, 2, 1, 0, 0, false)
	grid.AddItem(quoteBox, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(welcomeRibbon, 3, 0, 1, 2, 0, 0, false)

	Root.wonline.Title = titleBox

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

	loaderRibbon := ribbonPrimitive("CHESS-CLI -> Loading, please wait...", loaderMsgBox)

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

func (g *OnlineGame) Init() *cv.Grid {
	inputBox := cv.NewInputField()
	inputBox.SetDoneFunc(func(key tc.Key) {
		Root.gameState.NextMove = inputBox.GetText()
		inputBox.SetText("")

		if key == tc.KeyEnter {
			if contains(Root.gameState.LegalMoves, Root.gameState.NextMove) {
				err := g.DoMove(Root.gameState.NextMove)
				if err != nil {
					Root.gameState.Status += fmt.Sprintf("%v", err)
					Root.lgame.UpdateStatus()
				}
			} else {
				Root.gameState.Status += "Last input [red]invalid.[white]\n"
				g.UpdateStatus()
			}

		} else {
			Root.gameState.NextMove = ""
		}
	})
	inputBox.SetLabel("Enter your move: ")

	gameBox := boardPrimitive(g.OnlineTableHandler)

	statusBox := cv.NewTextView()
	statusBox.SetWordWrap(true)
	statusBox.SetDynamicColors(true)
	statusBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(inputBox)
		return nil
	})

	historyBox := cv.NewTextView()
	historyBox.SetWordWrap(true)
	historyBox.SetDynamicColors(true)
	historyBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(inputBox)
		return nil
	})

	userInfoBox := cv.NewTextView()
	userInfoBox.SetTextAlign(cv.AlignLeft)
	userInfoBox.SetVerticalAlign(cv.AlignTop)
	userInfoBox.SetDynamicColors(true)
	userInfoBox.SetText("USER INFO")
	userInfoBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(inputBox)
		return nil
	})

	userTimerBox := cv.NewTextView()
	userTimerBox.SetTextAlign(cv.AlignLeft)
	userTimerBox.SetDynamicColors(true)
	userTimerBox.SetText("∞")
	userTimerBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(inputBox)
		return nil
	})

	oppTimerBox := cv.NewTextView()
	oppTimerBox.SetTextAlign(cv.AlignLeft)
	oppTimerBox.SetDynamicColors(true)
	oppTimerBox.SetText("∞")
	oppTimerBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(inputBox)
		return nil
	})

	oppInfoBox := cv.NewTextView()
	oppInfoBox.SetTextAlign(cv.AlignLeft)
	oppInfoBox.SetVerticalAlign(cv.AlignBottom)
	oppInfoBox.SetDynamicColors(true)
	oppInfoBox.SetText("OPP INFO")
	oppInfoBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(inputBox)
		return nil
	})

	options := cv.NewList()
	optionsList := []string{"Back", "Quit"}
	optionsExplain := []string{"Go back Home", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoLichessFromGame, Root.App.Stop}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetSelectedFunc(optionsFunc[i])
		options.AddItem(item)
	}
	options.SetSelectedFocusOnly(true)
	options.SetHover(true)
	options.SetWrapAround(true)

	Ribbon := ribbonPrimitive(gameOnlineRibbonstr, inputBox)

	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -1)
	grid.SetRows(-1, 1, 1, -1, 12, 1)
	grid.SetBorders(false)

	//row, col, rowSpan, colSpain
	grid.AddItem(inputBox, 4, 1, 1, 1, 0, 0, true)
	grid.AddItem(Center(30, 10, gameBox, inputBox), 0, 1, 4, 1, 0, 0, false)
	grid.AddItem(statusBox, 0, 0, 2, 1, 0, 0, false)
	grid.AddItem(historyBox, 2, 0, 2, 1, 0, 0, false)
	grid.AddItem(Ribbon, 5, 0, 1, 3, 0, 0, false)

	grid.AddItem(oppInfoBox, 0, 2, 1, 1, 0, 0, false)
	grid.AddItem(oppTimerBox, 1, 2, 1, 1, 0, 0, false)
	grid.AddItem(userTimerBox, 2, 2, 1, 1, 0, 0, false)
	grid.AddItem(userInfoBox, 3, 2, 1, 1, 0, 0, false)
	grid.AddItem(options, 4, 0, 1, 1, 0, 0, false)

	g.Grid = grid
	g.Board = gameBox
	g.Status = statusBox
	g.History = historyBox
	g.OppTimer = oppTimerBox
	g.OppInfo = oppInfoBox
	g.UserTimer = userTimerBox
	g.UserInfo = userInfoBox
	g.List = options

	return grid
}

func (pg *OnlinePostGame) Init() *cv.Grid {
	postList := cv.NewList() //list construction
	postList.SetWrapAround(true)
	postList.SetHover(true)
	choices := []string{"Home", "New", "Ongoing", "Quit"}
	explain := []string{"Back to the welcome screen", "Create a new game", "Select from ongoing games", "Press to exit"}
	shortcuts := []rune{'h', 'n', 'o', 'q'}
	selectFunc := []ListSelectedFunc{gotoLichessAfterLogin, gotoChallengeConstruction, gotoOngoing, Root.App.Stop}

	for i := 0; i < len(choices); i++ {
		item := cv.NewListItem(choices[i])
		item.SetSecondaryText(explain[i])
		item.SetShortcut(rune(shortcuts[i]))
		item.SetSelectedFunc(selectFunc[i])
		postList.AddItem(item)
	}

	gameBox := cv.NewTable()
	gameBox.SetSelectable(false, false)
	gameBox.SetSortClicked(false)
	gameBox.SetFixed(10, 10)
	gameBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(postList)
		return nil
	})

	historyBox := cv.NewTextView()
	historyBox.SetWordWrap(true)
	historyBox.SetDynamicColors(true)
	historyBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(postList)
		return nil
	})

	resultBox := cv.NewTextView()
	resultBox.SetWordWrap(true)
	resultBox.SetDynamicColors(true)
	resultBox.SetTextAlign(cv.AlignCenter)
	resultBox.SetVerticalAlign(cv.AlignBottom)
	resultBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(postList)
		return nil
	})

	Ribbon := ribbonPrimitive(gameRibbonstr, postList)

	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -1)
	grid.SetRows(-1, -3, -1, 1)
	grid.SetBorders(false)

	//row col rowSpan colSpan
	grid.AddItem(postList, 1, 2, 1, 1, 0, 0, true)
	grid.AddItem(resultBox, 0, 0, 1, 3, 0, 0, false)
	grid.AddItem(Center(30, 10, gameBox, postList), 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(Ribbon, 3, 0, 1, 3, 0, 0, false)
	grid.AddItem(historyBox, 1, 0, 2, 1, 0, 0, false)

	pg.Grid = grid
	pg.Board = gameBox
	pg.Result = resultBox
	pg.History = historyBox

	return grid
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
	optionsList := []string{"Leave", "Quit"}
	optionsExplain := []string{"Go back Home", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoLichessAfterLogin, Root.App.Stop}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetSelectedFunc(optionsFunc[i])
		options.AddItem(item)
	}
	options.SetSelectedFocusOnly(true)
	options.SetHover(true)

	ribbon := ribbonPrimitive(OngoingRibbonstr, gameList)

	title := titlePrimitive("Select an ongoing game.", gameList)

	preview := boardPrimitive(func(row, col int) {})
	preview.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(gameList)
		return nil
	})

	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -4)
	grid.SetRows(10, -2, 10, 1)
	grid.SetBorders(false)

	//row col rowSpan colSpan
	grid.AddItem(gameList, 1, 2, 2, 1, 0, 0, true)
	grid.AddItem(title, 0, 0, 1, 3, 0, 0, false)
	grid.AddItem(Center(30, 10, preview, gameList), 1, 0, 1, 2, 0, 0, false)
	grid.AddItem(ribbon, 3, 0, 1, 3, 0, 0, false)
	grid.AddItem(options, 2, 1, 1, 1, 0, 0, false)

	ong.List = gameList
	ong.Preview = preview

	return grid
}

func (c *Challenges) Init() *cv.Grid {
	outgoing := cv.NewList()
	outgoing.AddItem(cv.NewListItem("placeholder out"))
	incoming := cv.NewList()
	incoming.AddItem(cv.NewListItem("placeholder in"))
	incoming.SetSelectedFunc(func(i int, li *cv.ListItem) {
		gameID := InChallengeGameID[i]
		for _, chal := range Root.User.IncomingChallenges {
			if chal.ID == gameID {
				currentGameID = gameID
				err := api.AcceptChallenge(gameID)
				if err != nil {
					gotoLichess()
				}
				startNewOnlineGame()
			}
		}

	})

	options := cv.NewList()
	options.SetSelectedFocusOnly(true)
	options.SetHover(true)
	optionsList := []string{"Back", "Quit"}
	optionsExplain := []string{"Go back Home", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoLichessAfterLogin, Root.App.Stop}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetSelectedFunc(optionsFunc[i])
		options.AddItem(item)
	}

	outtitle := titlePrimitive("Outgoing Challenges.", outgoing)
	intitle := titlePrimitive("Incoming Challenges.", incoming)
	ribbon := ribbonPrimitive("CHESS-CLI -> Challenges", incoming)

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

	return grid
}
