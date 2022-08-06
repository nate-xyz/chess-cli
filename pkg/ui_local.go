package pkg

import (
	"fmt"

	cv "code.rocketnine.space/tslocum/cview"
	tc "github.com/gdamore/tcell/v2"
)

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
	selectFunc := []ListSelectedFunc{startNewLocalGame, gotoLichess, Root.App.Stop}

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
	grid.SetBorders(false)

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
		Root.currentLocalGame.NextMove = inputBox.GetText()
		inputBox.SetText("")
		if key == tc.KeyEnter {
			if contains(Root.currentLocalGame.LegalMoves, Root.currentLocalGame.NextMove) {
				err := LocalGameDoMove(Root.currentLocalGame.NextMove)
				if err != nil {
					Root.currentLocalGame.Status += fmt.Sprintf("%v", err)
					UpdateGameStatus(Root.Status)
				}
			} else {
				Root.currentLocalGame.Status += "Last input [red]invalid.[white]\n"
				UpdateGameStatus(Root.Status)
			}
		} else {
			Root.currentLocalGame.NextMove = ""
		}
	})
	inputBox.SetLabel("Enter your move: ")

	options := cv.NewList()
	optionsList := []string{"Leave", "Quit"}
	optionsExplain := []string{"Go back Home", "Close chess-cli"}
	optionsFunc := []ListSelectedFunc{gotoWelcome, Root.App.Stop}
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

	Root.Board = gameBox
	Root.Status = statusBox
	Root.History = historyBox

	Root.InfoOppo = oppInfoBox
	Root.TimeOppo = oppTimerBox
	Root.TimeUser = userTimerBox
	Root.InfoUser = userInfoBox

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

	Root.PostBoard = gameBox
	Root.PostStatus = resultBox
	Root.PostHistory = historyBox

	return grid
}
