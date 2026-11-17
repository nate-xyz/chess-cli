package pkg

import (
	"context"
	"time"

	cv "codeberg.org/tslocum/cview"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
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
	Move      *moveBox

	engine      *Engine // set when playing a local Stockfish
	engineColor chess.Color
	engineLevel int
	engineGen   int // bumped when an engine game starts or stops
	clock       localClock
}

type PostGameScreen struct {
	*postGame
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
	Side      *cv.Flex
	Move      *moveBox
	Board     *BoardView
	Status    *cv.TextView
	History   *cv.TextView
	List      *cv.List
	UserInfo  *cv.TextView
	UserTimer *cv.TextView
	OppInfo   *cv.TextView
	OppTimer  *cv.TextView
	PopUp     *cv.Flex
	Chat      *cv.TextView
	ChatInput *cv.InputField

	Full  api.StreamBoardGameFull
	State api.StreamBoardGameState
	Gone  api.StreamOpponentGone

	// per-game state, reset by startNewOnlineGame; UI goroutine only
	ctx       context.Context
	cancel    context.CancelFunc
	ended     bool
	popupKind string    // which offer the popup is showing, if any
	goneSince time.Time // when the opponent-gone event arrived
	canClaim  bool      // opponent gone long enough to claim
	clockW    int64     // clocks (ms) as of clockAt
	clockB    int64
	clockAt   time.Time
	lowWarned bool // the low-time sound has played this game
}

type OnlinePostGame struct {
	*postGame
}

type Ongoing struct {
	List    *cv.List
	Preview *BoardView
}

type Challenges struct {
	Out *cv.List
	In  *cv.List
}
