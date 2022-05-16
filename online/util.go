package online

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/notnil/chess"
)

func MoveTranslation(sequence string) string {
	sequence_array := strings.Split(sequence, " ")
	game := chess.NewGame(chess.UseNotation(chess.UCINotation{}))

	for _, move := range sequence_array {
		if game.Outcome() == chess.NoOutcome {
			err := game.MoveStr(move)
			if err != nil {
				// handle error
				fmt.Printf("%v\n", err)
			}
			continue
		}
		//break
	}

	//

	if BoardFullGame.Black.Name == Username {
		return game.Position().Board().Flip(chess.UpDown).String()
	}
	return game.Position().String()
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

//writes to a local text file
func WriteLocal(title string, payload string) {
	f, err := os.Create(fmt.Sprintf("%s.txt", title))

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err2 := f.WriteString(payload)

	if err2 != nil {
		log.Fatal(err2)
	}
}
