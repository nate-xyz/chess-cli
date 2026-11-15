package pkg

import (
	cv "codeberg.org/tslocum/cview"
)

func InitUI() {
	lgame := new(GameScreen)
	pgame := new(PostGameScreen)
	wonline := new(WelcomeOnline)
	loader := new(Loader)
	ongame := new(OnlineGame)
	ponline := new(OnlinePostGame)
	ongoing := new(Ongoing)
	challenges := new(Challenges)
	sgame := new(SavedGames)
	Root.lgame = lgame
	Root.pgame = pgame
	Root.wonline = wonline
	Root.loader = loader
	Root.ongame = ongame
	Root.ponline = ponline
	Root.ongoing = ongoing
	Root.challenges = challenges
	Root.sgame = sgame
	panels := cv.NewPanels()
	panels.AddPanel("welcome", WelcomeInit(), true, true)
	panels.AddPanel("localgame", lgame.Init(), true, false)
	panels.AddPanel("postlocal", pgame.Init(), true, false)
	panels.AddPanel("lichesswelcome", wonline.Init(), true, false)
	panels.AddPanel("loader", loader.Init(), true, false)
	panels.AddPanel("onlinegame", ongame.Init(), true, false)
	panels.AddPanel("postonline", ponline.Init(), true, false)
	panels.AddPanel("challenge", initConstruct(), true, false)
	panels.AddPanel("ongoing", ongoing.Init(), true, false)
	panels.AddPanel("listchallenge", challenges.Init(), true, false)
	panels.AddPanel("saved", sgame.Init(), true, false)
	panels.AddPanel("options", optionsInit(), true, false)
	Root.puzzle = new(PuzzleScreen)
	panels.AddPanel("puzzle", Root.puzzle.Init(), true, false)
	Root.watch = new(WatchScreen)
	panels.AddPanel("watch", Root.watch.Init(), true, false)
	Root.study = new(StudyScreen)
	panels.AddPanel("study", Root.study.Init(), true, false)
	panels.AddPanel("prompt", promptInit(), true, false)
	setMoveBoxAlways(moveBoxAlways()) // apply the saved setting
	refreshNews()
	onboard := needsOnboarding()
	if onboard {
		panels.AddPanel("onboard", onboardInit(), true, false)
	}
	Root.nav = panels
	Root.sglist = new(SavedGameList)
	Root.sglist.Init()
	Root.App.SetRoot(panels, true)
	if onboard {
		Root.Switch("onboard")
		Root.App.SetFocus(onboardList)
	}
	Root.App.SetAfterDrawFunc(afterDraw)
	Root.App.SetMouseCapture(trackHover)
	Root.App.SetInputCapture(globalKeys)
}
