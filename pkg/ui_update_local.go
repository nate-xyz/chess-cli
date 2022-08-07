package pkg

import (
	"fmt"
	"strings"

	cv "code.rocketnine.space/tslocum/cview"
	"github.com/notnil/chess"
)

func (g *GameScreen) UpdateStatus() {
	var status string = Root.gameState.Status
	Root.gameState.Status = ""
	if Root.gameState.Game.Position().Turn() == chess.White {
		status += "White's turn. \n\n"
	} else {
		status += "Black's turn. \n\n"
	}

	Root.gameState.UpdateLegalMoves()
	status += "Valid Moves: [green]" + strings.Join(Root.gameState.LegalMoves, ", ") + "[white]\n"
	g.Status.SetText(status)
}

func (g *GameScreen) DoMove(m string) error {
	gClone := Root.gameState.Game.Clone()
	move, _ := GetMoveType(m, Root.gameState.Game) //store move
	p := GetPiece(m[2:], gClone).String()          //get piece
	err := Root.gameState.Game.MoveStr(m)          //do the move
	if err == nil {
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
		DrawBoard(g.Board, Root.gameState.Game.Position().Turn() == chess.White)
		Root.gameState.NextMove = "" //clear the next move
		g.UpdateStatus()
		g.UpdateUserInfo()
		// g.UpdateTime()
		Root.App.GetScreen().Beep()
		if Root.gameState.Game.Outcome() != chess.NoOutcome { //check if game is done
			gotoPostLocal()
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

	g.UserInfo.SetText(UserString)
	g.OppInfo.SetText(OppString)
}

func LocalTableHandler(row, col int) {
	selectedCell := translateSelectedCell(row, col, Root.gameState.Game.Position().Turn() == chess.White)
	if LastSelectedCell.Alg == selectedCell {
		//toggle selected status of this cell
		Root.lgame.Board.Select(100, 100)
		LastSelectedCell = PiecePosition{-1, -1, "", true, ""}
	} else {
		//try to do move
		todoMove := LastSelectedCell.Alg + selectedCell
		if contains(Root.gameState.LegalMoves, todoMove) {
			err := Root.lgame.DoMove(todoMove)
			if err != nil {
				Root.gameState.Status += fmt.Sprintf("%v", err)
				Root.lgame.UpdateStatus()
			}
		}
		//check if select is empty for updateBoard
		symbol := Root.lgame.Board.GetCell(row, col).GetText()
		LastSelectedCell = PiecePosition{row, col, selectedCell, (symbol == EmptyChar), symbol}
	}
	DrawBoard(Root.lgame.Board, Root.gameState.Game.Position().Turn() == chess.White)
}

func (g *PostGameScreen) UpdateResult() {
	g.Result.SetText(Root.gameState.Status)
	Root.gameState.Status = ""
}

func (sg *SavedGames) UpdateList() {
	sg.List.Clear()
	for i, game := range Root.sglist.Games {
		var text string
		if game.MoveCount%2 == 0 {
			text += " White to play"
		} else {
			text += "Black to play"
		}
		item := cv.NewListItem(text)
		item.SetSecondaryText(game.FEN)
		item.SetShortcut(rune('a' + i))
		sg.List.AddItem(item)
	}
}
