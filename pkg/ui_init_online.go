package pkg

import (
	"fmt"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
	"github.com/nate-xyz/chess-cli/api"
)

func (w *WelcomeOnline) Init() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -2, -1, 1)
	grid.SetBorders(false)

	titleBox := cv.NewTextView()
	titleBox.SetTextAlign(cv.AlignCenter)
	titleBox.SetVerticalAlign(cv.AlignMiddle)
	titleBox.SetDynamicColors(true)
	titleBox.SetText(LichessTitle)

	welcomeRibbon := ribbonPrimitive(LichessRibbon)
	quoteBox := quoutePrimitive(GetRandomQuote())
	welcomeList := cv.NewList()
	welcomeList.SetHover(true)
	welcomeList.SetWrapAround(true)
	choices := []string{"New Game", "Ongoing Games", "Challenges", "Back", "Quit", "Test Friend", "Test AI"}
	explain := []string{"Construct a new game request", "Select from your active games", "Incoming & outgoing challenges", "Back to welcome screen", "Press to exit", "aaaa", "bbbb"}
	shortcuts := []rune{'n', 'o', 'c', 'b', 'q', 'y', 'z'}
	selectFunc := []ListSelectedFunc{gotoChallengeConstruction, gotoOngoing, gotoChallenges, gotoWelcome, Root.App.Stop, TestFriend, TestAI}
	for i := 0; i < len(choices); i++ {
		item := cv.NewListItem(choices[i])
		item.SetSecondaryText(explain[i])
		item.SetShortcut(rune(shortcuts[i]))
		item.SetSelectedFunc(selectFunc[i])
		welcomeList.AddItem(item)
	}

	grid.AddItem(welcomeList, 1, 1, 1, 1, 0, 0, true)
	grid.AddItem(titleBox, 0, 0, 2, 1, 0, 0, false)
	grid.AddItem(quoteBox, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(welcomeRibbon, 3, 0, 1, 2, 0, 0, false)

	Root.wonline.Title = titleBox

	return grid

}

func (l *Loader) Init() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-1)
	grid.SetRows(-1, -1, 1)
	grid.SetBorders(false)

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

	loaderRibbon := ribbonPrimitive("CHESS-CLI -> Loading, please wait...")

	grid.AddItem(loaderMsgBox, 1, 0, 1, 1, 0, 0, true)
	grid.AddItem(loaderIconBox, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(loaderRibbon, 2, 0, 1, 1, 0, 0, false)

	l.Message = loaderMsgBox
	l.Icon = loaderIconBox

	return grid
}

func (g *OnlineGame) Init() *cv.Grid {
	gameBox := boardPrimitive(g.OnlineTableHandler)

	statusBox := cv.NewTextView()
	statusBox.SetWordWrap(true)
	statusBox.SetDynamicColors(true)

	historyBox := cv.NewTextView()
	historyBox.SetWordWrap(true)
	historyBox.SetDynamicColors(true)

	userInfoBox := cv.NewTextView()
	userInfoBox.SetTextAlign(cv.AlignLeft)
	userInfoBox.SetVerticalAlign(cv.AlignTop)
	userInfoBox.SetDynamicColors(true)
	userInfoBox.SetText("USER INFO")

	userTimerBox := cv.NewTextView()
	userTimerBox.SetTextAlign(cv.AlignLeft)
	userTimerBox.SetDynamicColors(true)
	userTimerBox.SetText("∞")

	oppTimerBox := cv.NewTextView()
	oppTimerBox.SetTextAlign(cv.AlignLeft)
	oppTimerBox.SetDynamicColors(true)
	oppTimerBox.SetText("∞")

	oppInfoBox := cv.NewTextView()
	oppInfoBox.SetTextAlign(cv.AlignLeft)
	oppInfoBox.SetVerticalAlign(cv.AlignBottom)
	oppInfoBox.SetDynamicColors(true)
	oppInfoBox.SetText("OPP INFO")

	//timerBox.SetText("TIME")

	inputBox := cv.NewInputField()

	inputBox.SetDoneFunc(func(key tc.Key) {
		Root.gameState.NextMove = inputBox.GetText()
		inputBox.SetText("")
		//TODO: print response status body to window if invalid
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

	Ribbon := ribbonPrimitive(gameOnlineRibbonstr)

	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -1)
	grid.SetRows(-1, 1, 1, -1, 12, 1)
	grid.SetBorders(false)

	//row, col, rowSpan, colSpain
	grid.AddItem(inputBox, 4, 1, 1, 1, 0, 0, true)
	grid.AddItem(Center(30, 10, gameBox), 0, 1, 4, 1, 0, 0, false)
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
	gameBox := cv.NewTable()
	gameBox.SetSelectable(false, false)
	gameBox.SetSortClicked(false)
	gameBox.SetFixed(10, 10)

	historyBox := cv.NewTextView()
	historyBox.SetWordWrap(true)
	historyBox.SetDynamicColors(true)

	resultBox := cv.NewTextView()
	resultBox.SetWordWrap(true)
	resultBox.SetDynamicColors(true)
	resultBox.SetTextAlign(cv.AlignCenter)
	resultBox.SetVerticalAlign(cv.AlignBottom)

	//list construction
	postList := cv.NewList()
	choices := []string{"Home", "New", "Quit"}
	explain := []string{"Back to the welcome screen", "Create a new game", "Press to exit"}
	shortcuts := []rune{'h', 'n', 'q'}

	selectFunc := []ListSelectedFunc{gotoLichessAfterLogin, gotoChallengeConstruction, Root.App.Stop}

	postList.SetWrapAround(true)
	postList.SetHover(true)

	for i := 0; i < len(choices); i++ {
		item := cv.NewListItem(choices[i])
		item.SetSecondaryText(explain[i])
		item.SetShortcut(rune(shortcuts[i]))
		item.SetSelectedFunc(selectFunc[i])
		postList.AddItem(item)
	}

	Ribbon := ribbonPrimitive(gameRibbonstr)

	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -1)
	grid.SetRows(-1, -3, -1, 1)
	grid.SetBorders(false)

	//row col rowSpan colSpan
	grid.AddItem(postList, 1, 2, 1, 1, 0, 0, true)
	grid.AddItem(resultBox, 0, 0, 1, 3, 0, 0, false)
	grid.AddItem(Center(30, 10, gameBox), 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(Ribbon, 3, 0, 1, 3, 0, 0, false)
	grid.AddItem(historyBox, 1, 0, 2, 1, 0, 0, false)

	pg.Grid = grid
	pg.Board = gameBox
	pg.Result = resultBox
	pg.History = historyBox

	return grid
}

func (ong *Ongoing) Init() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -4)
	grid.SetRows(10, -2, 10, 1)
	grid.SetBorders(false)

	preview := boardPrimitive(func(row, col int) {})

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

	ribbon := ribbonPrimitive(OngoingRibbonstr)

	title := cv.NewTextView()
	title.SetTextAlign(cv.AlignCenter)
	title.SetVerticalAlign(cv.AlignMiddle)
	title.SetDynamicColors(true)
	title.SetText("Select an ongoing game.")

	//row col rowSpan colSpan
	grid.AddItem(gameList, 1, 2, 2, 1, 0, 0, true)
	grid.AddItem(title, 0, 0, 1, 3, 0, 0, false)
	grid.AddItem(Center(30, 10, preview), 1, 0, 1, 2, 0, 0, false)
	grid.AddItem(ribbon, 3, 0, 1, 3, 0, 0, false)
	grid.AddItem(options, 2, 1, 1, 1, 0, 0, false)

	//Root.OngoingList = gameList

	ong.List = gameList
	ong.Preview = preview

	return grid
}

func (c *Challenges) Init() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-1, -3, -3, -1)
	grid.SetRows(-1, -3, 5, 1)
	grid.SetBorders(false)

	outtitle := cv.NewTextView()
	outtitle.SetTextAlign(cv.AlignCenter)
	outtitle.SetVerticalAlign(cv.AlignMiddle)
	outtitle.SetDynamicColors(true)
	outtitle.SetText("Outgoing Challenges.")

	intitle := cv.NewTextView()
	intitle.SetTextAlign(cv.AlignCenter)
	intitle.SetVerticalAlign(cv.AlignMiddle)
	intitle.SetDynamicColors(true)
	intitle.SetText("Incoming Challenges.")

	options := cv.NewList()
	optionsList := []string{"Back", "Quit"}
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

	outgoing := cv.NewList()
	outgoing.AddItem(cv.NewListItem("placeholder out"))
	incoming := cv.NewList()
	incoming.AddItem(cv.NewListItem("placeholder in"))
	incoming.SetSelectedFunc(func(i int, li *cv.ListItem) {
		gameID := InChallengeGameID[i]
		for _, chal := range Root.User.IncomingChallenges {
			if chal.Id == gameID {
				currentGameID = gameID
				err := api.AcceptChallenge(gameID)
				if err != nil {
					gotoLichess()
				}
				startNewOnlineGame()
			}
		}

	})
	ribbon := ribbonPrimitive("CHESS-CLI -> Challenges")

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
