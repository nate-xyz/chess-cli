package pkg

import "github.com/notnil/chess"

func (s *State) NewGame() {
	game := new(GameState)
	game.Init()
	s.gameState = game
}

func (s *State) Switch(panel string) {
	s.nav.SetCurrentPanel(panel)
}

func (s *State) RefreshAll() {
	s.App.QueueUpdateDraw(func() {})
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
