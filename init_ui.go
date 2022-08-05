package main

import (
	"fmt"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
)

func InitUI() {
	// bx := cv.NewBox()
	panels := cv.NewPanels()
	welcomeGrid := initWelcomeScreen()
	localGameGrid := initGameScreen()
	postLocalGrid := initPostGame()
	lichessWelcome := initWelcomeLichess()
	loader := initLoadingScreen()
	onlinegame := initLichessGameGrid()
	postonline := initPostOnline()
	constructchallenge := initConstruct()
	ongoing := initOngoing()
	challenges := initChallenges()
	panels.AddPanel("welcome", welcomeGrid, true, true)
	panels.AddPanel("localgame", localGameGrid, true, false)
	panels.AddPanel("postlocal", postLocalGrid, true, false)
	panels.AddPanel("lichesswelcome", lichessWelcome, true, false)
	panels.AddPanel("loader", loader, true, false)
	panels.AddPanel("onlinegame", onlinegame, true, false)
	panels.AddPanel("postonline", postonline, true, false)
	panels.AddPanel("challenge", constructchallenge, true, false)
	panels.AddPanel("ongoing", ongoing, true, false)
	panels.AddPanel("listchallenge", challenges, true, false)

	root.nav = panels
	root.app.SetRoot(panels, true)

}

func initWelcomeScreen() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -2, -1, 1)
	grid.SetBorders(false)

	titleBox := titlePrimitive(ApplicationTitle)

	welcomeRibbon := ribbonPrimitive(welcomeRibbonstr)
	quoteBox := quoutePrimitive(GetRandomQuote())

	//list construction
	welcomeList := cv.NewList()
	welcomeList.SetHover(true)
	welcomeList.SetWrapAround(true)

	choices := []string{"Local", "Online", "Quit"}
	explain := []string{"Play a local chess game", "Play a game on lichess", "Press to exit"}
	shortcuts := []rune{'l', 'o', 'q'}
	selectFunc := []ListSelectedFunc{startNewLocalGame, gotoLichess, root.app.Stop}

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

	return grid
}

//local game grid
func initGameScreen() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -1)
	grid.SetRows(-1, 1, 1, -1, 10, 1)
	grid.SetBorders(true)
	gameBox := boardPrimitive(LocalTableHandler)

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
	userInfoBox.SetText("[Blue]White")

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
	oppInfoBox.SetText("[red]Black")

	inputBox := cv.NewInputField()

	inputBox.SetDoneFunc(func(key tc.Key) {
		root.currentLocalGame.NextMove = inputBox.GetText()
		inputBox.SetText("")
		if key == tc.KeyEnter {
			if contains(root.currentLocalGame.LegalMoves, root.currentLocalGame.NextMove) {
				err := LocalGameDoMove(root.currentLocalGame.NextMove)
				if err != nil {
					root.currentLocalGame.Status += fmt.Sprintf("%v", err)
					UpdateGameStatus(root.Status)
				}
			} else {
				root.currentLocalGame.Status += "Last input [red]invalid.[white]\n"
				UpdateGameStatus(root.Status)
			}
		} else {
			root.currentLocalGame.NextMove = ""
		}
	})
	inputBox.SetLabel("Enter your move: ")

	options := cv.NewList()
	optionsList := []string{"Leave", "Quit"}
	optionsExplain := []string{"Go back Home", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoWelcome, root.app.Stop}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetSelectedFunc(optionsFunc[i])
		options.AddItem(item)
	}
	options.SetSelectedFocusOnly(true)
	options.SetHover(true)

	Ribbon := ribbonPrimitive(gameRibbonstr)

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

	root.Board = gameBox
	root.Status = statusBox
	root.History = historyBox

	root.InfoOppo = oppInfoBox
	root.TimeOppo = oppTimerBox
	root.TimeUser = userTimerBox
	root.InfoUser = userInfoBox

	return grid
}

func initPostGame() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -3, -1, 1)
	grid.SetBorders(false)

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

	//list construction
	postList := cv.NewList()
	postList.SetHover(true)
	choices := []string{"New", "Home", "Quit"}
	explain := []string{"Play a new game", "Back to the welcome screen", "Press to exit"}
	shortcuts := []rune{'a', 'b', 'q'}
	selectFunc := []ListSelectedFunc{startNewLocalGame, gotoWelcome, root.app.Stop}
	postList.SetWrapAround(true)
	for i := 0; i < len(choices); i++ {
		item := cv.NewListItem(choices[i])
		item.SetSecondaryText(explain[i])
		item.SetShortcut(rune(shortcuts[i]))
		item.SetSelectedFunc(selectFunc[i])
		postList.AddItem(item)
	}

	Ribbon := ribbonPrimitive(gameRibbonstr)

	grid.AddItem(postList, 2, 1, 1, 1, 0, 0, true)
	grid.AddItem(resultBox, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(Center(28, 10, gameBox), 1, 0, 2, 1, 0, 0, false)
	grid.AddItem(Ribbon, 3, 0, 1, 2, 0, 0, false)
	grid.AddItem(historyBox, 0, 1, 2, 1, 0, 0, false)

	root.PostBoard = gameBox
	root.PostStatus = resultBox
	root.PostHistory = historyBox

	return grid
}

func initLoadingScreen() *cv.Grid {
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

	root.LoaderIcon = loaderIconBox
	root.LoaderMsg = loaderMsgBox

	return grid
}

func initWelcomeLichess() *cv.Grid {
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
	selectFunc := []ListSelectedFunc{gotoChallengeConstruction, gotoOngoing, gotoChallenges, gotoWelcome, root.app.Stop, TestFriend, TestAI}
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

	root.LichessTitle = titleBox

	return grid

}

func initLichessGameGrid() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -1)
	grid.SetRows(-1, 1, 1, -1, 10, 1)
	grid.SetBorders(false)
	gameBox := boardPrimitive(OnlineTableHandler)

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
		root.currentLocalGame.NextMove = inputBox.GetText()
		inputBox.SetText("")
		//TODO: print response status body to window if invalid
		if key == tc.KeyEnter {
			if contains(root.currentLocalGame.LegalMoves, root.currentLocalGame.NextMove) {
				err := OnlineGameDoMove(root.currentLocalGame.NextMove)
				if err != nil {
					root.currentLocalGame.Status += fmt.Sprintf("%v", err)
					UpdateGameStatus(root.Status)
				}
			} else {
				root.currentLocalGame.Status += "Last input [red]invalid.[white]\n"
				UpdateGameStatus(root.OnlineStatus)
			}

		} else {
			root.currentLocalGame.NextMove = ""
		}
	})
	inputBox.SetLabel("Enter your move: ")

	options := cv.NewList()
	optionsList := []string{"Back", "Abort", "Offer Draw", "Resign", "Quit"}
	optionsExplain := []string{"Go back Home", "", "", "", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoLichess, doAbort, doNothing, doNothing, root.app.Stop}
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

	root.OnlineBoard = gameBox
	root.OnlineStatus = statusBox
	root.OnlineHistory = historyBox

	root.OnlineInfoOppo = oppInfoBox
	root.OnlineTimeOppo = oppTimerBox
	root.OnlineTimeUser = userTimerBox
	root.OnlineInfoUser = userInfoBox

	return grid
}

func initPostOnline() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -3, -1, 1)
	grid.SetBorders(false)

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

	//list construction
	postList := cv.NewList()
	choices := []string{"New", "Home", "Quit"}
	explain := []string{"Play a new game", "Back to the welcome screen", "Press to exit"}
	shortcuts := []rune{'a', 'b', 'q'}
	selectFunc := []ListSelectedFunc{startNewLocalGame, gotoWelcome, root.app.Stop}
	postList.SetWrapAround(true)
	for i := 0; i < len(choices); i++ {
		item := cv.NewListItem(choices[i])
		item.SetSecondaryText(explain[i])
		item.SetShortcut(rune(shortcuts[i]))
		item.SetSelectedFunc(selectFunc[i])
		postList.AddItem(item)
	}

	Ribbon := ribbonPrimitive(gameRibbonstr)

	grid.AddItem(postList, 2, 1, 1, 1, 0, 0, true)
	grid.AddItem(resultBox, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(Center(28, 10, gameBox), 1, 0, 2, 1, 0, 0, false)
	grid.AddItem(Ribbon, 3, 0, 1, 2, 0, 0, false)
	grid.AddItem(historyBox, 0, 1, 2, 1, 0, 0, false)

	root.OnlinePostBoard = gameBox
	root.OnlinePostStatus = resultBox
	root.OnlinePostHistory = historyBox

	return grid
}

func initOngoing() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -4)
	grid.SetRows(10, -2, 10, 1)
	grid.SetBorders(false)

	preview := boardPrimitive(func(row, col int) {})
	root.OngoingPreview = preview
	gameList := cv.NewList()
	gameList.SetHover(true)
	gameList.SetWrapAround(true)
	gameList.SetChangedFunc(func(i int, li *cv.ListItem) {
		gameID := GameListIDArr[i]
		for _, game := range OngoingGames {
			if game.FullID == gameID {
				FEN := game.Fen
				var white bool = (game.IsMyTurn && game.Color == "white") || (!game.IsMyTurn && game.Color == "black")
				FENtoBoard(root.OngoingPreview, FEN, white)
			}
		}
	})
	gameList.SetSelectedFunc(func(i int, li *cv.ListItem) {
		gameID := GameListIDArr[i]
		for _, game := range OngoingGames {
			if game.FullID == gameID {
				currentGameID = gameID
				startNewOnlineGame()
			}
		}

	})

	options := cv.NewList()
	optionsList := []string{"Leave", "Quit"}
	optionsExplain := []string{"Go back Home", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoLichess, root.app.Stop}
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
	root.OngoingList = gameList

	return grid

}

func initChallenges() *cv.Grid {
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
	optionsFunc := []ListSelectedFunc{gotoLichess, root.app.Stop}
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
		for _, chal := range IncomingChallenges {
			if chal.Id == gameID {
				currentGameID = gameID
				err := AcceptChallenge(gameID)
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

	root.OutChallengeList = outgoing
	root.InChallengeList = incoming

	return grid

}
