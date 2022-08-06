package pkg

import (
	cv "code.rocketnine.space/tslocum/cview"
)

func InitUI() {
	// bx := cv.NewBox()
	panels := cv.NewPanels()
	panels.AddPanel("welcome", initWelcomeScreen(), true, true)
	panels.AddPanel("localgame", initGameScreen(), true, false)
	panels.AddPanel("postlocal", initPostGame(), true, false)
	panels.AddPanel("lichesswelcome", initWelcomeLichess(), true, false)
	panels.AddPanel("loader", initLoadingScreen(), true, false)
	panels.AddPanel("onlinegame", initLichessGameGrid(), true, false)
	panels.AddPanel("postonline", initPostOnline(), true, false)
	panels.AddPanel("challenge", initConstruct(), true, false)
	panels.AddPanel("ongoing", initOngoing(), true, false)
	panels.AddPanel("listchallenge", initChallenges(), true, false)

	Root.nav = panels
	Root.App.SetRoot(panels, true)

}
