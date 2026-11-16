package pkg

import (
	"fmt"
	"strings"

	cv "codeberg.org/tslocum/cview"
	"github.com/notnil/chess"
)

func (g *GameScreen) UpdateStatus() {
	var status string = Root.gameState.Status
	Root.gameState.Status = ""
	if Root.gameState.Game.Position().Turn() == chess.White {
		status = "White to move. " + status
	} else {
		status = "Black to move. " + status
	}
	if g.engineTurn() {
		status += "\n[#a0a0a0]Stockfish is thinking…[-]"
	}
	g.Status.SetText(status)
}

func (g *GameScreen) DoMove(m string) error {
	gClone := Root.gameState.Game.Clone()
	move, _ := GetMoveType(m, Root.gameState.Game) //store move
	p := GetPiece(m[2:], gClone).String()          //get piece
	mover := Root.gameState.Game.Position().Turn()
	err := Root.gameState.Game.MoveStr(m) //do the move
	if err == nil {
		g.clockMoved(mover)
		if move.HasTag(chess.Check) {
			Root.gameState.Status += "Check!"
		}
		if move.HasTag(chess.Capture) {
			if len(Root.gameState.MoveHistoryArray)%2 == 0 { //white
				Root.gameState.WhiteCaptured = append(Root.gameState.WhiteCaptured, p)
			} else {
				Root.gameState.BlackCaptured = append(Root.gameState.BlackCaptured, p)
			}
		}
		Root.gameState.MoveHistoryArray = append(Root.gameState.MoveHistoryArray, m)
		DrawMoveHistory(g.History)
		DrawBoard(g.Board, g.whiteAtBottom())
		Root.gameState.NextMove = "" //clear the next move
		g.UpdateStatus()
		g.UpdateUserInfo()
		// g.UpdateTime()
		Root.App.GetScreen().Beep()
		if Root.gameState.Game.Outcome() != chess.NoOutcome { //check if game is done
			gotoPostLocal()
		} else if g.engineTurn() {
			g.engineMove()
		}
		return nil
	}
	return err
}

func (g *GameScreen) UpdateUserInfo() {
	var (
		OppName      string
		UserString   string = "[blue]%v[white]"
		OppString    string = "[red]%v[white]"
		BlackCapture string = strings.Join(Root.gameState.BlackCaptured, "") + " \t"
		WhiteCapture string = strings.Join(Root.gameState.WhiteCaptured, "") + " \t"
	)
	var PlayerName string
	if Root.gameState.Game.Position().Turn() == chess.White {
		OppName = "Black"
		PlayerName = "White"
		UserString = fmt.Sprintf(UserString, PlayerName)
		OppString = fmt.Sprintf(OppString, OppName)
		UserString = fmt.Sprintf("%v\n%v", UserString, WhiteCapture)
		OppString = fmt.Sprintf("%v\n%v", BlackCapture, OppString)
	} else {
		PlayerName = "Black"
		OppName = "White"
		UserString = fmt.Sprintf(UserString, PlayerName)
		OppString = fmt.Sprintf(OppString, OppName)
		UserString = fmt.Sprintf("%v\n%v", UserString, BlackCapture)
		OppString = fmt.Sprintf("%v\n%v", WhiteCapture, OppString)
	}

	if g.vsEngine() {
		you, sf := WhiteCapture, BlackCapture
		youName, sfName := "White", "Black"
		if g.engineColor == chess.White {
			you, sf, youName, sfName = BlackCapture, WhiteCapture, "Black", "White"
		}
		UserString = fmt.Sprintf("[blue]You (%s)[white]\n%v", youName, you)
		OppString = fmt.Sprintf("%v\n[red]Stockfish level %d (%s)[white]", sf, g.engineLevel, sfName)
	}
	g.UserInfo.SetText(UserString)
	g.OppInfo.SetText(OppString)
}

// LocalSquareHandler handles a click (or Enter) on a board square in a local game.
func LocalSquareHandler(sq chess.Square) {
	g := Root.lgame
	g.Board.HandleSquare(Root.gameState.Game, sq, !g.engineTurn(), func(uci string) {
		if err := g.DoMove(uci); err != nil {
			Root.gameState.Status += fmt.Sprintf("%v", err)
			g.UpdateStatus()
		}
	})
	DrawBoard(g.Board, g.whiteAtBottom())
}

func (sg *SavedGames) UpdateList() {
	sg.List.Clear()
	for i, game := range Root.sglist.Games {
		var text string
		if game.MoveCount%2 == 0 {
			text += "White's Turn. "
		} else {
			text += "Black's Turn. "
		}

		if game.LastPlayed == game.Started {
			text += "Created: [" + game.Started + "]"
		} else {
			text += "Played: [" + game.LastPlayed + "] "
			text += "Created: [" + game.Started + "] "
		}

		item := cv.NewListItem(text)
		item.SetSecondaryText(game.FEN)
		item.SetShortcut(numberKey(i))
		sg.List.AddItem(item)
	}
}
