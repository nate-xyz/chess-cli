package pkg

import (
	cv "code.rocketnine.space/tslocum/cview"
	"github.com/nate-xyz/chess-cli/api"
)

type GameScreen struct {
	Board     *BoardView
	Status    *cv.TextView
	History   *cv.TextView
	UserInfo  *cv.TextView
	UserTimer *cv.TextView
	OppInfo   *cv.TextView
	OppTimer  *cv.TextView
	List      *cv.List
}

type PostGameScreen struct {
	Board   *BoardView
	Result  *cv.TextView
	History *cv.TextView
}

type SavedGames struct {
	List    *cv.List
	Preview *BoardView
}

type WelcomeOnline struct {
	Title *cv.TextView
}

type Loader struct {
	Message *cv.TextView
	Icon    *cv.TextView
}

type OnlineGame struct {
	Grid      *cv.Grid
	Board     *BoardView
	Status    *cv.TextView
	History   *cv.TextView
	List      *cv.List
	UserInfo  *cv.TextView
	UserTimer *cv.TextView
	OppInfo   *cv.TextView
	OppTimer  *cv.TextView
	PopUp     *cv.Flex

	Full   api.StreamBoardGameFull
	State  api.StreamBoardGameState
	Resign api.StreamBoardResign
}

type OnlinePostGame struct {
	Grid *cv.Grid

	Board   *BoardView
	Result  *cv.TextView
	History *cv.TextView
	PopUp   *cv.Flex
}

type Ongoing struct {
	List    *cv.List
	Preview *BoardView
}

type Challenges struct {
	Out *cv.List
	In  *cv.List
}
