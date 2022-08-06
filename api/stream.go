package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

//starts event stream for a user after login
func StreamEvent(EventChannel chan<- StreamEventType, got_token chan struct{}) error {
	<-got_token
	StreamEventStarted = true
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/stream/event", HostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Content-Type", "application/x-ndjson")
	if err != nil {
		//WriteLocal("STREAM EVENT REQ ERR", fmt.Sprintf("stream event get request failed: %v", err))
		log.Fatal(err)
	}
	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(fmt.Errorf("%v: %v", resp.StatusCode, err))
		return fmt.Errorf("%v: %v", resp.StatusCode, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		//WriteLocal("STREAM EVENT RESP ERR", fmt.Sprintf("bad response: %v", resp.StatusCode))
		log.Fatal(fmt.Errorf("bad response: %v", resp.StatusCode))
	}

	Online = true
	// go func() {
	// 	UpdateLichessTitle("")
	// }()

	d := json.NewDecoder(resp.Body)

	for {
		select {
		// case <-quit_stream:
		// 	return nil
		default:
			var responseData map[string]interface{}
			err := d.Decode(&responseData)
			if err != nil {
				if err != io.EOF {
					return err
				}
				continue
			}
			if !isNil(responseData["type"]) {

				streamEvent := responseData["type"].(string)
				var streamEventID string
				var streamSource string
				switch streamEvent {
				case "gameStart", "gameFinish": //game type stream event
					game := responseData["game"].(map[string]interface{})
					streamEventID = game["id"].(string)
					streamSource = game["source"].(string)

					jsonStr, err := json.Marshal(responseData)
					if err != nil {
						log.Fatal(err)
					}
					if err := json.Unmarshal(jsonStr, &CurrentStreamEventGame); err != nil {
						log.Fatal(err)
					}

				case "challenge", "challengeCanceled", "challengeDeclined": //challenge type stream event
					challenge := responseData["challenge"].(map[string]interface{})
					streamEventID = challenge["id"].(string)

					jsonStr, err := json.Marshal(responseData)
					if err != nil {
						log.Fatal(err)
					}
					if err := json.Unmarshal(jsonStr, &CurrentStreamEventChallenge); err != nil {
						log.Fatal(err)
					}

				}
				EventChannel <- StreamEventType{streamEvent, streamEventID, streamSource}
			} else {
				Online = false
				// go func() {
				// 	UpdateLichessTitle("")
				// }()
				return fmt.Errorf("invalid stream event")
			}
		}
	}
}

//consumes stream events for a user after stream for a user started after login
func StreamConsumer(EventChannel <-chan StreamEventType) {
	for {
		select {
		case e := <-EventChannel:
			//fmt.Printf("consumer: %v %v \n", e.Event, e.Id)
			EventStreamArr = append([]StreamEventType{e}, EventStreamArr...)
			//StreamChannelForWaiter <- e

			// WriteLocal(fmt.Sprintf("StreamEvent_%v", time.Now().Format("2006-01-02_15:04:05")),
			// 	fmt.Sprintf("%v\n%v", e.Id, e.Event))

			if e.EventType != "gameStart" {
				// noti <- fmt.Sprintf("event %v:%v", e.Event, e.Id)
			} else {
				if !containedInOngoingGames(OngoingGames, e.GameID) {
					// noti <- fmt.Sprintf("new game! %v:%v", e.Event, e.Id)
				}
			}
			//noti <- fmt.Sprintf("event %v:%v", e.Event, e.Id)
			//CheckGlobal = e.Id

			//TODO: add to global data structure storing all events
		}
	}
}

//after initializing a game, this function streams the the state of the board of the game
func StreamBoardState(EventChannel chan<- BoardEvent, StreamError chan<- error, game string) {
	//https://lichess.org/api/board/game/stream/{gameId}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/board/game/stream/%s", HostUrl, game), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Content-Type", "application/x-ndjson")
	if err != nil {
		StreamError <- fmt.Errorf("stream event get request failed: %v", err)
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		StreamError <- fmt.Errorf("%v: %v", resp.StatusCode, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {

		StreamError <- fmt.Errorf("bad response: %v", resp.StatusCode)
	}

	d := json.NewDecoder(resp.Body)

	for {
		var responseData map[string]interface{}
		err := d.Decode(&responseData)

		if err != nil {
			if err != io.EOF {
				StreamError <- err
				//close(StreamError)
				return
			} else {
				EventChannel <- EOF
				//close(EventChannel)
				return
			}
		}

		if !isNil(responseData["type"]) { //make sure response is valid
			streamEvent := responseData["type"].(string)
			var bevent BoardEvent
			// bevent.Type = streamEvent
			// bevent.Moves = ""
			switch streamEvent {
			case "gameFull":
				bevent = GameFull
				jsonStr, err := json.Marshal(responseData)
				if err != nil {
					log.Fatal(err)
				}
				if err := json.Unmarshal(jsonStr, &BoardFullGame); err != nil {
					log.Fatal(err)
				}
			case "gameState":
				bevent = GameState
				jsonStr, err := json.Marshal(responseData)
				if err != nil {
					log.Fatal(err)
				}
				if err := json.Unmarshal(jsonStr, &BoardGameState); err != nil {
					log.Fatal(err)
				}

			case "chatLine":
				//TODO:
				bevent = ChatLine

			case "chatLineSpectator":
				//TODO:
				bevent = ChatLineSpectator

			case "gameStateResign":
				bevent = GameStateResign
				jsonStr, err := json.Marshal(responseData)
				if err != nil {
					log.Fatal(err)
				}
				if err := json.Unmarshal(jsonStr, &BoardResign); err != nil {
					log.Fatal(err)
				}

			}

			//signal what event just happened so LichessGameScreen knows what global to look at
			EventChannel <- bevent
		} else {
			StreamError <- fmt.Errorf("invalid ndjson")
			return
		}
	}
}

//consumes board events after the board stream has been started for a game

// func BoardConsumer(EventChannel <-chan BoardEvent) {
// 	for {
// 		select {
// 		case e := <-EventChannel:
// 			//fmt.Printf("consumer: %v %v \n", e.Event, e.Id)
// 			BoardStreamArr = append([]BoardState{e}, BoardStreamArr...)
// 			WriteLocal(fmt.Sprintf("BoardEvent_%v", time.Now().Format("2006-01-02_15:04:05")),
// 				fmt.Sprintf("%v\n%v\n%v\n%v", e.Type, e.Moves, e.Status, e.Rated))
// 		}
// 	}
// }
