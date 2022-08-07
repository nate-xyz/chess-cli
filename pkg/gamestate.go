package pkg

import (
	cv "code.rocketnine.space/tslocum/cview"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func (s *State) NewGame() {
	game := new(GameState)
	game.Init()
	s.gameState = game
}

func (s *State) Login() error {
	s.User = new(Login)
	err := s.User.Init()
	if err != nil {
		return err
	}
	return nil
}

func (user *Login) Init() error {
	var err error
	user.Token, err = api.PerformOAuth()
	if err != nil {
		return err
	}
	err = user.GetLichessUserInfo()
	if err != nil {
		return err
	}
	return nil
}

func (s *State) Switch(panel string) {
	s.nav.SetCurrentPanel(panel)
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
