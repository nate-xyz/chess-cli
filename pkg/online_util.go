package pkg

import (
	"strings"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

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

func containedInEventStream(a []api.StreamEventType, gameid string) (string, bool) {
	for _, e := range a {
		if e.GameID == gameid {
			return e.EventType, true
		}
	}
	return "", false
}

func EventContainedInEventStream(a []api.StreamEventType, eventtype string) (api.StreamEventType, bool) {
	for _, e := range a {
		if e.EventType == eventtype {
			return e, true
		}
	}
	return api.StreamEventType{}, false
}

func getEvents(a []api.StreamEventType, gameid string) ([]api.StreamEventType, bool) {
	n := make([]api.StreamEventType, 0)
	for _, e := range a {
		if e.GameID == gameid {
			n = append(n, e)
		}
	}
	if len(n) > 0 {
		return n, true
	}
	return n, false
}

// func GameOutcome(sequence string) (string, string) {
// 	sequence_array := strings.Split(sequence, " ")
// 	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))

// 	for _, move := range sequence_array {
// 		if game.Outcome() == chess.NoOutcome {
// 			err := game.MoveStr(move)
// 			if err != nil {
// 				// handle error
// 				//fmt.Printf("%v\n", err)
// 				return "", ""
// 			}
// 			continue
// 		}
// 	}
// 	outcome_str := fmt.Sprintf("Game completed. %s by %s.", game.Outcome(), game.Method())
// 	var name_str string
// 	if game.Outcome()[0] == '1' {
// 		name_str = fmt.Sprintf("White (%s) wins.", api.BoardFullGame.White.Name)
// 	}
// 	if game.Outcome()[0] == '0' {
// 		name_str = fmt.Sprintf("Black (%s) wins.", api.BoardFullGame.Black.Name)
// 	}

// 	return outcome_str, name_str
// }
