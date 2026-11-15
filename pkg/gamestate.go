package pkg

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"

	cv "codeberg.org/tslocum/cview"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func (s *State) NewGame() {
	game := new(GameState)
	game.Init()
	s.gameState = game
}

// Init logs in (reusing a saved token if it's still valid) and loads the
// account details. showURL is called with the browser login URL, if needed.
func (user *Login) Init(ctx context.Context, showURL func(string)) error {
	var err error
	user.Token, err = api.PerformOAuth(ctx, showURL)
	if err != nil {
		return err
	}
	err = user.GetLichessUserInfo()
	var apiErr *api.APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
		// the saved token was revoked: log in again
		if err := api.ClearToken(); err != nil {
			return err
		}
		if user.Token, err = api.AuthUser(ctx, showURL); err != nil {
			return err
		}
		err = user.GetLichessUserInfo()
	}
	return err
}

// frontPanel is the screen showing, readable from any goroutine (background
// tickers use it to stay idle when their screen isn't open).
var frontPanel atomic.Value

func showing(panel string) bool {
	p, _ := frontPanel.Load().(string)
	return p == panel
}

func (s *State) Switch(panel string) {
	frontPanel.Store(panel)
	s.nav.SetCurrentPanel(panel)
	if panel == "onlinegame" || panel == "watch" {
		applyLayoutPrefs()
	}
	focusMoveBox(panel)
}

func (s *State) RefreshAll(p ...cv.Primitive) {
	s.App.QueueUpdateDraw(func() {}, p...)
}

func (s *GameState) Init() {
	s.Game = chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	s.NextMove = ""
	s.MoveHistoryArray = nil
	s.LegalMoves = nil
	s.Status = ""
	s.MoveCount = 0
	s.WhiteCaptured = nil
	s.BlackCaptured = nil
}

func (gs *GameState) UpdateLegalMoves() {
	gs.LegalMoves = []string{}
	for _, move := range gs.Game.ValidMoves() {
		gs.LegalMoves = append(gs.LegalMoves, move.String())
	}
}
