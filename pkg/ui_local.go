package pkg

import (
	"fmt"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
)

func WelcomeInit() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -2, -1, 1)
	grid.SetBorders(false)

	titleBox := titlePrimitive(ApplicationTitle)

	welcomeRibbon := ribbonPrimitive(welcomeRibbonstr)

	quoteBox := quoutePrimitive(GetRandomQuote())

	//list construction
	List := cv.NewList()
	List.SetHover(true)
	List.SetWrapAround(true)

	choices := []string{"Local", "Online", "Quit"}
	explain := []string{"Play a local chess game", "Play a game on lichess", "Press to exit"}
	shortcuts := []rune{'l', 'o', 'q'}
	selectFunc := []ListSelectedFunc{startNewLocalGame, gotoLichess, Root.App.Stop}

	for i := 0; i < len(choices); i++ {
		item := cv.NewListItem(choices[i])
		item.SetSecondaryText(explain[i])
		item.SetShortcut(rune(shortcuts[i]))
		item.SetSelectedFunc(selectFunc[i])
		List.AddItem(item)
	}

	grid.AddItem(List, 1, 1, 1, 1, 0, 0, true)
	grid.AddItem(titleBox, 0, 0, 2, 1, 0, 0, false)
	grid.AddItem(quoteBox, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(welcomeRibbon, 3, 0, 1, 2, 0, 0, false)

	return grid
}

//local game grid
func (g *GameScreen) Init() *cv.Grid {
	grid := cv.NewGrid()
	grid.SetColumns(-1, -2, -1)
	grid.SetRows(-1, 1, 1, -1, 10, 1)
	grid.SetBorders(false)

	g.Board = boardPrimitive(LocalTableHandler)

	g.Status = cv.NewTextView()
	g.Status.SetWordWrap(true)
	g.Status.SetDynamicColors(true)

	g.History = cv.NewTextView()
	g.History.SetWordWrap(true)
	g.History.SetDynamicColors(true)

	g.UserInfo = cv.NewTextView()
	g.UserInfo.SetTextAlign(cv.AlignLeft)
	g.UserInfo.SetVerticalAlign(cv.AlignTop)
	g.UserInfo.SetDynamicColors(true)
	g.UserInfo.SetText("[Blue]White")

	g.UserTimer = cv.NewTextView()
	g.UserTimer.SetTextAlign(cv.AlignLeft)
	g.UserTimer.SetDynamicColors(true)
	g.UserTimer.SetText("∞")

	g.OppTimer = cv.NewTextView()
	g.OppTimer.SetTextAlign(cv.AlignLeft)
	g.OppTimer.SetDynamicColors(true)
	g.OppTimer.SetText("∞")

	g.OppInfo = cv.NewTextView()
	g.OppInfo.SetTextAlign(cv.AlignLeft)
	g.OppInfo.SetVerticalAlign(cv.AlignBottom)
	g.OppInfo.SetDynamicColors(true)
	g.OppInfo.SetText("[red]Black")

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

	g.List = cv.NewList()
	optionsList := []string{"Leave", "Quit"}
	optionsExplain := []string{"Go back Home", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoWelcome, Root.App.Stop}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSecondaryText(optionsExplain[i])
		item.SetSelectedFunc(optionsFunc[i])
		g.List.AddItem(item)
	}
	g.List.SetSelectedFocusOnly(true)
	g.List.SetHover(true)

	Ribbon := ribbonPrimitive(gameRibbonstr)

	grid.AddItem(Input, 4, 1, 1, 1, 0, 0, true)
	grid.AddItem(Center(30, 10, g.Board), 0, 1, 4, 1, 0, 0, false)
	grid.AddItem(g.Status, 0, 0, 2, 1, 0, 0, false)
	grid.AddItem(g.History, 2, 0, 2, 1, 0, 0, false)
	grid.AddItem(Ribbon, 5, 0, 1, 3, 0, 0, false)

	grid.AddItem(g.OppInfo, 0, 2, 1, 1, 0, 0, false)
	grid.AddItem(g.OppTimer, 1, 2, 1, 1, 0, 0, false)
	grid.AddItem(g.UserTimer, 2, 2, 1, 1, 0, 0, false)
	grid.AddItem(g.UserInfo, 3, 2, 1, 1, 0, 0, false)

	grid.AddItem(g.List, 4, 0, 1, 1, 0, 0, false)

	// Root.Board = g.Board
	// Root.Status = g.Status
	// Root.History = g.History
	// Root.InfoOppo = g.OppInfo
	// Root.TimeOppo = g.OppTimer
	// Root.TimeUser = g.UserTimer
	// Root.InfoUser = g.UserInfo

	return grid

}

func (p *PostGameScreen) Init() *cv.Grid {
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
	selectFunc := []ListSelectedFunc{startNewLocalGame, gotoWelcome, Root.App.Stop}
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

	p.Board = gameBox
	p.Result = resultBox
	p.History = historyBox

	return grid
}
