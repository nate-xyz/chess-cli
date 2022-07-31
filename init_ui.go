package main

import (
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
	panels.AddPanel("welcome", welcomeGrid, true, true)
	panels.AddPanel("localgame", localGameGrid, true, false)
	panels.AddPanel("postlocal", postLocalGrid, true, false)
	panels.AddPanel("lichesswelcome", lichessWelcome, true, false)
	panels.AddPanel("loader", loader, true, false)
	panels.AddPanel("onlinegame", onlinegame, true, false)
	panels.AddPanel("postonline", postonline, true, false)
	panels.AddPanel("challenge", constructchallenge, true, false)

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
	choices := []string{"Local", "Online", "Stockfish", "Quit"}
	explain := []string{"Play a local chess game", "Play a game on lichess", "Play a game against AI", "Press to exit"}
	shortcuts := []rune{'a', 'b', 'c', 'q'}
	selectFunc := []ListSelectedFunc{startNewLocalGame, gotoLichess, doNothing, root.app.Stop}
	welcomeList.SetWrapAround(true)
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

func initGameScreen() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -1, 1, 1)
	grid.SetBorders(false)
	gameBox := boardPrimitive()

	statusBox := cv.NewTextView()
	statusBox.SetWordWrap(true)
	statusBox.SetDynamicColors(true)
	historyBox := cv.NewTextView()
	historyBox.SetWordWrap(true)
	historyBox.SetDynamicColors(true)

	inputBox := cv.NewInputField()

	inputBox.SetDoneFunc(func(key tc.Key) {
		root.currentLocalGame.NextMove = inputBox.GetText()
		inputBox.SetText("")
		if key == tc.KeyEnter {

			if contains(root.currentLocalGame.LegalMoves, root.currentLocalGame.NextMove) {
				LocalGameDoMove()
			} else {
				root.currentLocalGame.Status += "Last input [red]invalid.[white]\n"
				UpdateGameStatus(root.Status)
			}
		} else {
			root.currentLocalGame.NextMove = ""
		}
	})
	inputBox.SetLabel("Enter your move: ")

	Ribbon := ribbonPrimitive(gameRibbonstr)

	grid.AddItem(inputBox, 2, 0, 1, 1, 0, 0, true)
	grid.AddItem(Center(28, 10, gameBox), 0, 0, 2, 1, 0, 0, false)
	grid.AddItem(statusBox, 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(historyBox, 1, 1, 2, 1, 0, 0, false)
	grid.AddItem(Ribbon, 3, 0, 1, 2, 0, 0, false)

	root.Board = gameBox
	root.Status = statusBox
	root.History = historyBox

	return grid
}

func initPostGame() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -3, -1, 1)
	grid.SetBorders(false)
	gameBox := boardPrimitive()

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
	loaderIconBox.SetVerticalAlign(cv.AlignCenter)
	loaderMsgBox := cv.NewTextView()
	loaderMsgBox.SetWordWrap(true)
	loaderMsgBox.SetDynamicColors(true)
	loaderMsgBox.SetTextAlign(cv.AlignCenter)
	loaderMsgBox.SetVerticalAlign(cv.AlignCenter)
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
	titleBox.SetVerticalAlign(cv.AlignCenter)
	titleBox.SetDynamicColors(true)
	titleBox.SetText(LichessTitle)

	welcomeRibbon := ribbonPrimitive(LichessRibbon)
	quoteBox := quoutePrimitive(GetRandomQuote())
	welcomeList := cv.NewList()
	welcomeList.SetWrapAround(true)
	choices := []string{"New Game", "Ongoing Games", "Back", "Quit", "Test Friend", "Test AI"}
	explain := []string{"Construct a new game request", "Select from your active games", "Back to welcome screen", "Press to exit", "aaaa", "bbbb"}
	shortcuts := []rune{'n', 'o', 'b', 'q', 'y', 'z'}
	selectFunc := []ListSelectedFunc{gotoChallengeConstruction, doNothing, gotoWelcome, root.app.Stop, TestFriend, doNothing}
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
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -1, 1, 1)
	grid.SetBorders(false)
	gameBox := boardPrimitive()

	statusBox := cv.NewTextView()
	statusBox.SetWordWrap(true)
	statusBox.SetDynamicColors(true)
	historyBox := cv.NewTextView()
	historyBox.SetWordWrap(true)
	historyBox.SetDynamicColors(true)

	inputBox := cv.NewInputField()

	inputBox.SetDoneFunc(func(key tc.Key) {
		root.currentLocalGame.NextMove = inputBox.GetText()
		inputBox.SetText("")
		//TODO: print response status body to window if invalid
		if key == tc.KeyEnter {
			if contains(root.currentLocalGame.LegalMoves, root.currentLocalGame.NextMove) {
				OnlineGameDoMove()
			} else {
				root.currentLocalGame.Status += "Last input [red]invalid.[white]\n"
				UpdateGameStatus(root.OnlineStatus)
			}
		} else {
			root.currentLocalGame.NextMove = ""
		}
	})
	inputBox.SetLabel("Enter your move: ")

	Ribbon := ribbonPrimitive(gameOnlineRibbonstr)

	grid.AddItem(inputBox, 2, 0, 1, 1, 0, 0, true)
	grid.AddItem(Center(28, 10, gameBox), 0, 0, 2, 1, 0, 0, false)
	grid.AddItem(statusBox, 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(historyBox, 1, 1, 2, 1, 0, 0, false)
	grid.AddItem(Ribbon, 3, 0, 1, 2, 0, 0, false)

	root.OnlineBoard = gameBox
	root.OnlineStatus = statusBox
	root.OnlineHistory = historyBox

	return grid
}

func initPostOnline() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -3, -1, 1)
	grid.SetBorders(false)
	gameBox := boardPrimitive()

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
