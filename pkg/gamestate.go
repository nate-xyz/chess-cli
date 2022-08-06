package pkg

func (s *State) NewGame() {
	game := new(LocalGame)
	game.Init()
	s.gameState = game
}

func (s *State) Switch(panel string) {
	s.nav.SetCurrentPanel(panel)
}

func (s *State) RefreshAll() {
	s.App.QueueUpdateDraw(func() {})
}

func (gs *LocalGame) UpdateLegalMoves() {
	gs.LegalMoves = []string{}
	for _, move := range gs.Game.ValidMoves() {
		gs.LegalMoves = append(gs.LegalMoves, move.String())
	}
}
