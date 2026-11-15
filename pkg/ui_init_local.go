package pkg

import (
	"fmt"
	"strings"

	"github.com/nate-xyz/chess-cli/api"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
)

// resetWelcome puts the home screen back to its main menu.
var resetWelcome = func() {}

func WelcomeInit() *cv.Grid {
	List := cv.NewList()
	List.SetHover(true)
	List.SetWrapAround(true)

	//list construction
	var secondList, puzzleList, watchList, studyList func()
	var mainList func()

	mainList = func() {
		List.Clear()
		online := api.HasValidToken() // Watch and Studies only once you've logged in to Lichess
		for _, c := range []struct {
			text, explain string
			key           rune
			f             ListSelectedFunc
			needsLogin    bool
		}{
			{"Local", "Play a local chess game", 'l', secondList, false},
			{"Online", "Play a game on lichess", 'o', gotoLichess, false},
			{"Puzzles", "Lichess puzzles: daily or training, rated or casual", 'p', puzzleList, false},
			{"Watch", "Lichess TV, a player's game, or any game", 'w', watchList, true},
			{"Studies", "Your Lichess studies, or any study by link", 't', studyList, true},
			{"Options", "How the board and pieces look", 's', gotoOptions, false},
			{"Quit", "Press to exit", 'q', Root.App.Stop, false},
		} {
			if c.needsLogin && !online {
				continue
			}
			item := cv.NewListItem(c.text)
			item.SetSecondaryText(c.explain)
			item.SetShortcut(c.key)
			item.SetSelectedFunc(c.f)
			List.AddItem(item)
		}
	}

	secondList = func() {
		List.Clear()
		standard := cv.NewListItem("New")
		standard.SetShortcut('n')
		standard.SetSecondaryText("Two players, one keyboard")
		standard.SetSelectedFunc(startNewLocalGame)

		engine := cv.NewListItem("Play local Stockfish")
		engine.SetShortcut('f')
		engine.SetSecondaryText("Play the Stockfish engine on this computer, offline")
		engine.SetSelectedFunc(func() { stockfishList(List, secondList) })

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
		clock := cv.NewListItem("Clock: " + localClockName(localClockIndex()))
		clock.SetShortcut('c')
		clock.SetSecondaryText("Time control for local games; Enter for the next")
		clock.SetSelectedFunc(func() {
			nextLocalClock()
			cur := List.GetCurrentItemIndex()
			secondList()
			List.SetCurrentItem(cur)
		})
		List.AddItem(standard)
		List.AddItem(engine)
		List.AddItem(clock)
		List.AddItem(ongoing)
		List.AddItem(back)
		List.AddItem(quit)
	}

	puzzleList = func() {
		cur := List.GetCurrentItemIndex()
		List.Clear()
		add := func(text, explain string, key rune, f func()) {
			item := cv.NewListItem(text)
			item.SetSecondaryText(explain)
			item.SetShortcut(key)
			item.SetSelectedFunc(f)
			List.AddItem(item)
		}
		add("Daily puzzle", "Today's Lichess puzzle", 'd', func() { startPuzzle(true) })
		add("Puzzle training", "New puzzles, one after another", 'p', func() { startPuzzle(false) })
		mode, other := "Rated", "casual: no rating change, no login needed"
		if !puzzleRated() {
			mode, other = "Casual", "rated: changes your Lichess puzzle rating"
		}
		add("Mode: "+mode, "Switch to "+other, 'r', func() {
			savePuzzlePrefs(puzzleRated(), puzzleDifficulty())
			puzzleList()
		})
		add("Difficulty: "+puzzleDifficulty(), "Relative to your puzzle rating; Enter for the next level", 'f', func() {
			d := api.PuzzleDifficulties
			next := d[0]
			for i, x := range d {
				if x == puzzleDifficulty() {
					next = d[(i+1)%len(d)]
				}
			}
			savePuzzlePrefs(!puzzleRated(), next)
			puzzleList()
		})
		add("Back", "Go back to main options", 'b', mainList)
		List.SetCurrentItem(cur)
	}

	watchList = func() {
		List.Clear()
		add := func(text, explain string, key rune, f func()) {
			item := cv.NewListItem(text)
			item.SetSecondaryText(explain)
			item.SetShortcut(key)
			item.SetSelectedFunc(f)
			List.AddItem(item)
		}
		keys := "tuzrc9om" // best, bullet, blitz, rapid, classical, 960, bots, computer
		for i, ch := range api.TvChannels {
			ch := ch
			explain := "The best " + strings.ToLower(ch.Name) + " game on Lichess right now"
			if ch.Key == "best" {
				explain = "Lichess TV: the top-rated game being played"
			}
			add("TV: "+ch.Name, explain, rune(keys[i]), func() { WatchTv(ch.Key, ch.Name) })
		}
		add("Watch a player…", "Follow someone's current game", 'p', func() { askWatch("player") })
		add("Watch a game…", "Paste a Lichess game link or ID", 'g', func() { askWatch("game") })
		add("Back", "Go back to main options", 'b', mainList)
	}

	studyList = func() {
		List.Clear()
		for _, c := range []struct {
			text, explain string
			key           rune
			f             func()
		}{
			{"My studies", "The studies on your Lichess account, private ones too", 'm', func() { loadMyStudies(List, studyList) }},
			{"Open a study…", "Paste a Lichess study link or ID", 'o', askStudy},
			{"Back", "Go back to main options", 'b', mainList},
		} {
			item := cv.NewListItem(c.text)
			item.SetSecondaryText(c.explain)
			item.SetShortcut(c.key)
			item.SetSelectedFunc(c.f)
			List.AddItem(item)
		}
	}

	mainList()
	resetWelcome = mainList

	titleBox := titlePrimitive(ApplicationTitle, List)
	welcomeRibbon := ribbonPrimitive(welcomeRibbonstr, List)
	quoteBox := quoutePrimitive(GetRandomQuote(), List)

	grid := cv.NewGrid()
	grid.SetColumns(-2, -1)
	grid.SetRows(-1, -2, -1, 1)
	grid.SetBorders(false)

	newsBox = textBox(cv.AlignLeft)
	newsBox.SetPadding(0, 0, 4, 2)
	newsBox.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		Root.App.SetFocus(List)
		return nil
	})

	grid.AddItem(List, 1, 1, 1, 1, 0, 0, true)
	grid.AddItem(titleBox, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(newsBox, 1, 0, 1, 1, 0, 0, false)
	grid.AddItem(quoteBox, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(welcomeRibbon, 3, 0, 1, 2, 0, 0, false)

	return grid
}

// local game screen
func (g *GameScreen) Init() cv.Primitive {
	Board := NewBoardView(LocalSquareHandler)
	Board.EnableSounds()

	List := menuList()
	optionsList := []string{"Leave game", "Save", "Save as new", "Quit"}
	optionsFunc := []ListSelectedFunc{gotoWelcome, func() { doSave(false) }, func() { doSave(true) }, Root.App.Stop}
	for i, opt := range optionsList {
		item := cv.NewListItem(opt)
		item.SetSelectedFunc(optionsFunc[i])
		List.AddItem(item)
	}

	OppInfo := textBox(cv.AlignLeft)
	OppInfo.SetText("[red]Black")
	OppTimer := textBox(cv.AlignLeft)
	Status := textBox(cv.AlignLeft)
	History := textBox(cv.AlignLeft)
	UserTimer := textBox(cv.AlignLeft)
	UserInfo := textBox(cv.AlignLeft)
	UserInfo.SetText("[Blue]White")

	side := cv.NewFlex()
	side.SetDirection(cv.FlexRow)
	g.Move = newMoveBox(side, Board, func(text string) {
		if g.engineTurn() {
			Root.gameState.Status += "Wait for Stockfish's move. "
			g.UpdateStatus()
			return
		}
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
	side.AddItem(OppInfo, 2, 0, false)
	side.AddItem(OppTimer, 1, 0, false)
	side.AddItem(Status, 0, 1, false)
	side.AddItem(History, 0, 2, false)
	side.AddItem(g.Move.input, 0, 0, false)
	side.AddItem(List, len(optionsList), 0, false)
	side.AddItem(UserTimer, 1, 0, false)
	side.AddItem(UserInfo, 2, 0, false)

	// Tab: board <-> menu
	Board.SetTabFunc(func() { Root.App.SetFocus(List) })
	List.SetInputCapture(func(event *tc.EventKey) *tc.EventKey {
		if event.Key() == tc.KeyTab || event.Key() == tc.KeyEscape {
			Root.App.SetFocus(Board)
			return nil
		}
		return event
	})

	go g.clockTicker()
	screen := newGameLayout(Board, side, gameRibbonstr)

	g.Board = Board
	g.Status = Status
	g.History = History
	linkMoveList(History, Board)
	g.UserInfo = UserInfo
	g.UserTimer = UserTimer
	g.OppInfo = OppInfo
	g.OppTimer = OppTimer
	g.List = List

	return screen
}

func (p *PostGameScreen) Init() cv.Primitive {
	pg, screen := newPostGame([]*cv.ListItem{
		menuItem("New game", 'n', startNewLocalGame),
		menuItem("Saved games", 's', gotoSaved),
		menuItem("Home", 'h', gotoWelcome),
		menuItem("Quit", 'q', Root.App.Stop),
	}, postRibbonstr)
	p.postGame = pg
	return screen
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

	preview := NewBoardView(nil)
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
	grid.AddItem(preview, 1, 0, 1, 2, 0, 0, false)
	grid.AddItem(ribbon, 3, 0, 1, 3, 0, 0, false)
	grid.AddItem(options, 2, 1, 1, 1, 0, 0, false)

	sg.List = gameList
	sg.Preview = preview

	return grid
}
