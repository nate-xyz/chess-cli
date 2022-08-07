package pkg

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/notnil/chess"
)

type SavedInfo struct {
	FEN              string   `json:"fen"`
	MoveCount        int      `json:"move_count"`
	MoveHistoryArray []string `json:"move_history_array"`
	WhiteCaptured    []string `json:"white_captured"`
	BlackCaptured    []string `json:"black_captured"`
}

type SavedGameList struct {
	Games []SavedInfo `json:"games"`
}

func doSave() {
	Root.gameState.SaveGame()
	Root.Switch("welcome")
}

func (s *SavedGameList) Init() {
	checkForSavedGames()
}

func (gs *GameState) SaveGame() {
	var saved SavedInfo = SavedInfo{
		FEN:              gs.Game.Position().String(),
		MoveCount:        gs.MoveCount,
		MoveHistoryArray: gs.MoveHistoryArray,
		WhiteCaptured:    gs.WhiteCaptured,
		BlackCaptured:    gs.BlackCaptured,
	}
	Root.sglist.Games = append(Root.sglist.Games, saved)
	writeToSavedGames()
}

func RestoreGame(info SavedInfo) error {
	Root.NewGame()
	Root.gameState.MoveCount = info.MoveCount
	Root.gameState.MoveHistoryArray = info.MoveHistoryArray
	Root.gameState.WhiteCaptured = info.WhiteCaptured
	Root.gameState.BlackCaptured = info.BlackCaptured
	fen, err := chess.FEN(info.FEN)
	if err != nil {
		return err
	}
	Root.gameState.Game = chess.NewGame(fen)
	DrawBoard(Root.lgame.Board, Root.gameState.Game.Position().Turn() == chess.White)
	DrawMoveHistory(Root.lgame.History)
	Root.lgame.UpdateStatus()
	Root.Switch("localgame")
	return nil
}

func checkForSavedGames() error {
	_, err := os.Stat(saved_path)
	if err == nil {
		jsonFile, err := os.Open(saved_path)
		if err != nil {
			return err
		}
		defer jsonFile.Close()
		byteValue, _ := io.ReadAll(jsonFile)
		err = json.Unmarshal(byteValue, &Root.sglist)
		if err != nil {
			return err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist

		b, err := json.Marshal(&Root.sglist)
		if err != nil {
			return err
		}
		err = os.WriteFile(saved_path, b, 0644)
		if err != nil {
			return err
		}
	} else {
		return err // file may or may not exist. See err for details.
	}
	return nil
}

func writeToSavedGames() error {
	list := Root.sglist
	b, err := json.Marshal(*list)
	if err != nil {
		return err
	}
	err = os.WriteFile(saved_path, b, 0644)
	if err != nil {
		return err
	}
	return nil
}
