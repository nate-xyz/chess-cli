package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

type SavedInfo struct {
	FEN              string   `json:"fen"`
	MoveCount        int      `json:"move_count"`
	MoveHistoryArray []string `json:"move_history_array"`
	WhiteCaptured    []string `json:"white_captured"`
	BlackCaptured    []string `json:"black_captured"`
	Hash             string   `json:"hash"`
	LastPlayed       string   `json:"date"`
	Started          string   `json:"started"`
}

type SavedGameList struct {
	Games []SavedInfo `json:"games"`
}

func doSave(b bool) {
	Root.gameState.SaveGame(b)
	Root.Switch("welcome")
}

func (s *SavedGameList) Init() {
	checkForSavedGames()
}

func RemoveIndex(s []SavedInfo, index int) []SavedInfo {
	return append(s[:index], s[index+1:]...)
}

func (gs *GameState) SaveGame(new bool) {
	current_time := time.Now()
	var saved SavedInfo = SavedInfo{
		FEN:              gs.Game.Position().String(),
		MoveCount:        gs.MoveCount,
		MoveHistoryArray: gs.MoveHistoryArray,
		WhiteCaptured:    gs.WhiteCaptured,
		BlackCaptured:    gs.BlackCaptured,
		Started:          gs.Started,
		Hash:             gs.Hash,
		LastPlayed:       current_time.Format("06-01-02 15:04:05"),
	}
	if gs.Started == "" {
		saved.Started = current_time.Format("06-01-02 15:04:05")
	}
	if gs.Hash == "" { //unnamed game
		saved.Hash = RandStringRunes(12)
	} else if !new { //save over previous game
		for i, game := range Root.sglist.Games {
			if game.Hash == gs.Hash {
				Root.sglist.Games = RemoveIndex(Root.sglist.Games, i)
				break
			}
		}
	}
	Root.sglist.Games = append(Root.sglist.Games, saved)
	if err := writeToSavedGames(); err != nil {
		Root.gameState.Status += fmt.Sprintf("[red]Could not save: %v[-]\n", err)
	}
}

func RestoreGame(info SavedInfo) error {
	Root.NewGame()
	Root.gameState.MoveCount = info.MoveCount
	Root.gameState.MoveHistoryArray = info.MoveHistoryArray
	Root.gameState.WhiteCaptured = info.WhiteCaptured
	Root.gameState.BlackCaptured = info.BlackCaptured
	Root.gameState.Started = info.Started
	Root.gameState.Hash = info.Hash
	Root.gameState.LastPlayed = info.LastPlayed
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

// savedGamesPath is ~/.config/chess-cli/saved_games.json (or the platform
// equivalent). Older versions saved to ./saved_games.json; if the new file
// doesn't exist yet, that file is copied over (and left in place).
func savedGamesPath() (string, error) {
	dir, err := api.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "saved_games.json"), nil
}

func checkForSavedGames() error {
	path, err := savedGamesPath()
	if err != nil {
		return err
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		b, err = os.ReadFile(saved_path) // legacy location
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err == nil {
			if err := json.Unmarshal(b, &Root.sglist); err != nil {
				return err
			}
			return writeToSavedGames()
		}
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &Root.sglist)
}

func writeToSavedGames() error {
	path, err := savedGamesPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(Root.sglist, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
