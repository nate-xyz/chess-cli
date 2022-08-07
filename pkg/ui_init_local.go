package pkg

import (
	"fmt"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
)

func WelcomeInit() *cv.Grid {
	List := cv.NewList()
	List.SetHover(true)
	List.SetWrapAround(true)

	//list construction
	var secondList func()

	mainList := func() {
		List.Clear()
		choices := []string{"Local", "Online", "Quit"}
		explain := []string{"Play a local chess game", "Play a game on lichess", "Press to exit"}
		shortcuts := []rune{'l', 'o', 'q'}
		selectFunc := []ListSelectedFunc{secondList, gotoLichess, Root.App.Stop}

		for i := 0; i < len(choices); i++ {
			item := cv.NewListItem(choices[i])
			item.SetSecondaryText(explain[i])
			item.SetShortcut(rune(shortcuts[i]))
			item.SetSelectedFunc(selectFunc[i])
			List.AddItem(item)
		}
	}

	secondList = func() {
		List.Clear()
		standard := cv.NewListItem("New")
		standard.SetShortcut('n')
		standard.SetSecondaryText("Begin a new game")
		standard.SetSelectedFunc(startNewLocalGame)

		ongoing := cv.NewListItem("Saved")
		ongoing.SetSecondaryText("Select a game from your saved games")
		ongoing.SetShortcut('s')
		ongoing.SetSelectedFunc(gotoSaved)

		back := cv.NewListItem("Back")
		back.SetShortcut('b')
		back.SetSecondaryText("Go back to main options")
		back.SetSelectedFunc(mainList)

		quit := cv.NewListItem("Quit")
		quit.SetShortcut('q')
		quit.SetSecondaryText("Press to exit")
		quit.SetSelectedFunc(Root.App.Stop)
		List.AddItem(standard)
		List.AddItem(ongoing)
		List.AddItem(back)
		List.AddItem(quit)
	}

	mainList()

	titleBox := titlePrimitive(ApplicationTitle, List)
	welcomeRibbon := ribbonPrimitive(welcomeRibbonstr, List)
	quoteBox := quoutePrimitive(GetRandomQuote(), List)

	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -2, -1, 1)
	grid.SetBorders(false)

	grid.AddItem(List, 1, 1, 1, 1, 0, 0, true)
	grid.AddItem(titleBox, 0, 0, 2, 1, 0, 0, false)
	grid.AddItem(quoteBox, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(welcomeRibbon, 3, 0, 1, 2, 0, 0, false)

	return grid
}

//local game grid
func (g *GameScreen) Init() *cv.Grid {
	Input := cv.NewInputField()
	Input.SetDoneFunc(func(key tc.Key) {
		Root.gameState.NextMove = Input.GetText()
		Input.SetText("")
		if key == tc.KeyEnter {
			if contains(Root.gameState.LegalMoves, Root.gameState.NextMove) {
				err := g.DoMove(Root.gameState.NextMove)
				if err != nil {
					Root.gameState.Status += fmt.Sprintf("%v", err)
					Root.lgame.UpdateStatus()
				}
			} else {
				Root.gameState.Status += "Last input [red]invalid.[white]\n"
				Root.lgame.UpdateStatus()
			}
		} else {
			Root.gameState.NextMove = ""
		}
	})
	Input.SetLabel("Enter your move: ")

	List := cv.NewList()
	List.SetSelectedFocusOnly(true)
	List.SetHover(true)
	optionsList := []string{"Leave", "Save", "Save as new", "Quit"}
	optionsExplain := []string{"Go back Home", "Save this game locally", "Save without rewrite", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoWelcome, func() { doSave(false) }, func() { doSave(true) }, Root.App.Stop}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetSelectedFunc(optionsFunc[i])
		List.AddItem(item)
	}

	Board := boardPrimitive(LocalTableHandler)

	Status := cv.NewTextView()
	Status.SetWordWrap(true)
	Status.SetDynamicColors(true)
	Status.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(Input)
		return nil
	})

	History := cv.NewTextView()
	History.SetWordWrap(true)
	History.SetDynamicColors(true)
	History.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(Input)
		return nil
	})

	UserInfo := cv.NewTextView()
	UserInfo.SetTextAlign(cv.AlignLeft)
	UserInfo.SetVerticalAlign(cv.AlignTop)
	UserInfo.SetDynamicColors(true)
	UserInfo.SetText("[Blue]White")
	UserInfo.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(Input)
		return nil
	})

	UserTimer := cv.NewTextView()
	UserTimer.SetTextAlign(cv.AlignLeft)
	UserTimer.SetDynamicColors(true)
	UserTimer.SetText("∞")
	UserTimer.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(Input)
		return nil
	})

	OppTimer := cv.NewTextView()
	OppTimer.SetTextAlign(cv.AlignLeft)
	OppTimer.SetDynamicColors(true)
	OppTimer.SetText("∞")
	OppTimer.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(Input)
		return nil
	})

	OppInfo := cv.NewTextView()
	OppInfo.SetTextAlign(cv.AlignLeft)
	OppInfo.SetVerticalAlign(cv.AlignBottom)
	OppInfo.SetDynamicColors(true)
	OppInfo.SetText("[red]Black")
	OppInfo.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(Input)
		return nil
	})

	Ribbon := ribbonPrimitive(gameRibbonstr, Input)

	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -1)
	grid.SetRows(-1, 1, 1, -1, 10, 1)
	grid.SetBorders(false)

	grid.AddItem(Input, 4, 1, 1, 1, 0, 0, true)
	grid.AddItem(Center(30, 10, Board, Input), 0, 1, 4, 1, 0, 0, false)
	grid.AddItem(Status, 0, 0, 2, 1, 0, 0, false)
	grid.AddItem(History, 2, 0, 2, 1, 0, 0, false)
	grid.AddItem(Ribbon, 5, 0, 1, 3, 0, 0, false)
	grid.AddItem(OppInfo, 0, 2, 1, 1, 0, 0, false)
	grid.AddItem(OppTimer, 1, 2, 1, 1, 0, 0, false)
	grid.AddItem(UserTimer, 2, 2, 1, 1, 0, 0, false)
	grid.AddItem(UserInfo, 3, 2, 1, 1, 0, 0, false)
	grid.AddItem(List, 4, 0, 1, 1, 0, 0, false)

	g.Board = Board
	g.Status = Status
	g.History = History
	g.UserInfo = UserInfo
	g.UserTimer = UserTimer
	g.OppInfo = OppInfo
	g.OppTimer = OppTimer
	g.List = List

	return grid
}

func (p *PostGameScreen) Init() *cv.Grid {
	postList := cv.NewList() //list construction
	postList.SetWrapAround(true)
	postList.SetHover(true)
	choices := []string{"New", "Ongoing", "Home", "Quit"}
	explain := []string{"Play a new game", "Select from your saved games", "Back to the welcome screen", "Press to exit"}
	shortcuts := []rune{'n', 'o', 'h', 'q'}
	selectFunc := []ListSelectedFunc{startNewLocalGame, gotoSaved, gotoWelcome, Root.App.Stop}
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
	resultBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(postList)
		return nil
	})

	Ribbon := ribbonPrimitive(gameRibbonstr, postList)

	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -3, -1, 1)
	grid.SetBorders(false)

	grid.AddItem(postList, 2, 1, 1, 1, 0, 0, true)
	grid.AddItem(resultBox, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(Center(28, 10, gameBox, postList), 1, 0, 2, 1, 0, 0, false)
	grid.AddItem(Ribbon, 3, 0, 1, 2, 0, 0, false)
	grid.AddItem(historyBox, 0, 1, 2, 1, 0, 0, false)

	p.Board = gameBox
	p.Result = resultBox
	p.History = historyBox

	return grid
}

func (sg *SavedGames) Init() *cv.Grid {
	gameList := cv.NewList()
	gameList.SetHover(false)
	gameList.SetWrapAround(true)
	gameList.SetChangedFunc(func(i int, li *cv.ListItem) {
		game := Root.sglist.Games[i]
		FENtoBoard(sg.Preview, game.FEN, game.MoveCount%2 == 0)
	})
	gameList.SetSelectedFunc(func(i int, li *cv.ListItem) {
		game := Root.sglist.Games[i]
		err := RestoreGame(game)
		if err != nil {
			gotoWelcome()
		}
	})

	options := cv.NewList()
	options.SetSelectedFocusOnly(true)
	options.SetWrapAround(true)
	options.SetHover(true)
	optionsList := []string{"Leave", "Quit"}
	optionsExplain := []string{"Go back Home", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoWelcome, Root.App.Stop}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetSelectedFunc(optionsFunc[i])
		options.AddItem(item)
	}

	title := titlePrimitive("Select a saved game.", gameList)

	preview := boardPrimitive(func(row, col int) {})
	preview.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(gameList)
		return nil
	})

	ribbon := ribbonPrimitive(OngoingRibbonstr, gameList)

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

	sg.List = gameList
	sg.Preview = preview

	return grid
}
