package main

import (
	"fmt"
	"strings"

	"github.com/notnil/chess"
)

func GameOutcome(sequence string) (string, string) {
	sequence_array := strings.Split(sequence, " ")
	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))

	for _, move := range sequence_array {
		if game.Outcome() == chess.NoOutcome {
			err := game.MoveStr(move)
			if err != nil {
				// handle error
				//fmt.Printf("%v\n", err)
				return "", ""
			}
			continue
		}
	}
	outcome_str := fmt.Sprintf("Game completed. %s by %s.", game.Outcome(), game.Method())
	var name_str string
	if game.Outcome()[0] == '1' {
		name_str = fmt.Sprintf("White (%s) wins.", BoardFullGame.White.Name)
	}
	if game.Outcome()[0] == '0' {
		name_str = fmt.Sprintf("Black (%s) wins.", BoardFullGame.Black.Name)
	}

	return outcome_str, name_str
}

func MoveTranslationToFEN(sequence string) (string, error) {
	sequence_array := strings.Split(sequence, " ")
	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))
	if sequence != "" {
		for _, move := range sequence_array {
			if game.Outcome() == chess.NoOutcome {
				err := game.MoveStr(move)
				if err != nil {
					return "", err // handle error
				}
				continue
			}
		}
	}
	return game.Position().String(), nil
}

func containedInOngoingGames(a []OngoingGameInfo, gameid string) bool {
	for _, g := range a {
		if g.GameID == gameid {
			return true
		}
	}
	return false
}

func containedInEventStream(a []StreamEventType, gameid string) (string, bool) {
	for _, e := range a {
		if e.Id == gameid {
			return e.Event, true
		}
	}
	return "", false
}

func getEvents(a []StreamEventType, gameid string) ([]StreamEventType, bool) {
	n := make([]StreamEventType, 0)
	for _, e := range a {
		if e.Id == gameid {
			n = append(n, e)
		}
	}
	if len(n) > 0 {
		return n, true
	}
	return n, false
}
