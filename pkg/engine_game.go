package pkg

import (
	"context"
	"fmt"
	"time"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

// Playing a local Stockfish: the local game screen with the board turned to
// your side; Stockfish replies on its own.

func (g *GameScreen) vsEngine() bool { return g.engine != nil }

// engineTurn reports whether it's Stockfish's move.
func (g *GameScreen) engineTurn() bool {
	return g.engine != nil && Root.gameState.Game.Position().Turn() == g.engineColor
}

// whiteAtBottom: against Stockfish your side stays at the bottom; two-player
// games turn the board to whoever is moving.
func (g *GameScreen) whiteAtBottom() bool {
	if g.engine != nil {
		return g.engineColor == chess.Black
	}
	return Root.gameState.Game.Position().Turn() == chess.White
}

// stopEngine ends the Stockfish game's engine, if any.
func (g *GameScreen) stopEngine() {
	if g.engine != nil {
		e := g.engine
		g.engine = nil
		g.engineGen++
		go e.Close()
	}
}

// findEngine is the Stockfish to use: the one chosen in chess-cli, one on the
// system, or one chess-cli downloaded. "" if there's none.
func findEngine() string {
	if p := api.UserInfo.StockfishPath; p != "" {
		return p
	}
	if p := FindStockfish(); p != "" {
		return p
	}
	return downloadedStockfish()
}

// checkEngine tries an engine, so a bad path is caught when it's chosen.
func checkEngine(path string) error {
	e, err := StartEngine(path, 1)
	if err != nil {
		return err
	}
	e.Close()
	return nil
}

// startEngineGame starts a game against Stockfish at a level (1-8), with
// the player as player.
func startEngineGame(path string, level int, player chess.Color) {
	g := Root.lgame
	g.stopEngine()
	Root.NewGame()
	g.startClock()
	g.tickClock()
	g.engineColor = player.Other()
	g.engineLevel = level
	g.engineGen++
	gen := g.engineGen
	Root.gameState.Status = "Starting Stockfish…"
	Root.Switch("localgame")
	playSound(soundStart)
	DrawBoard(g.Board, player == chess.White)
	g.UpdateStatus()
	g.Board.ClearSelection()
	go func() {
		e, err := StartEngine(path, level)
		Root.App.QueueUpdateDraw(func() {
			if gen != g.engineGen {
				if e != nil {
					go e.Close()
				}
				return
			}
			if err != nil {
				Root.gameState.Status = fmt.Sprintf("[red]Couldn't start Stockfish: %v[-]\n", err)
				g.UpdateStatus()
				return
			}
			g.engine = e
			g.UpdateUserInfo()
			g.UpdateStatus()
			if g.engineTurn() {
				g.engineMove()
			}
		})
	}()
}

// engineMove asks Stockfish for its move and plays it.
func (g *GameScreen) engineMove() {
	e, gen := g.engine, g.engineGen
	game := Root.gameState.Game
	start := game.Positions()[0].String()
	var moves []string
	for i, m := range game.Moves() {
		moves = append(moves, chess.UCINotation{}.Encode(game.Positions()[i], m))
	}
	g.UpdateStatus()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		uci, err := e.BestMove(ctx, start, moves)
		Root.App.QueueUpdateDraw(func() {
			if gen != g.engineGen || g.engine != e {
				return // the game was left or restarted meanwhile
			}
			if err != nil {
				Root.gameState.Status = fmt.Sprintf("[red]Stockfish: %v[-]\n", err)
				g.UpdateStatus()
				return
			}
			if err := g.DoMove(uci); err != nil {
				Root.gameState.Status = fmt.Sprintf("[red]Stockfish played an illegal move (%s): %v[-]\n", uci, err)
				g.UpdateStatus()
			}
		})
	}()
}
