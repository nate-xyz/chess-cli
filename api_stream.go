package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

//starts event stream for a user after login
func StreamEvent(event_chan chan<- StreamEventType, got_token chan struct{}) error {
	<-got_token
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/stream/event", hostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Accept", "application/x-ndjson")
	if err != nil {
		WriteLocal("STREAM EVENT REQ ERR", fmt.Sprintf("stream event get request failed: %v", err))
		log.Fatal(err)
		os.Exit(1)
	}
	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("%v: %v", resp.StatusCode, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		WriteLocal("STREAM EVENT RESP ERR", fmt.Sprintf("bad response: %v", resp.StatusCode))
		log.Fatal(err)
		os.Exit(1)
	}
	Online = true
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
				event_chan <- StreamEventType{streamEvent, streamEventID, streamSource}
			} else {
				Online = false
				return fmt.Errorf("invalid stream event")
			}
		}
	}
}

//consumes stream events for a user after stream for a user started after login
func StreamConsumer(event_chan <-chan StreamEventType) {
	for {
		select {
		case e := <-event_chan:
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
//func StreamBoardState(event_chan chan<- BoardState, game string, ErrorMessage chan<- error, move_seq chan<- string) error {
func StreamBoardState(event_chan chan<- BoardEvent, game string) error {
	//https://lichess.org/api/board/game/stream/{gameId}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/board/game/stream/%s", hostUrl, game), nil)
	//NotiMessage <- fmt.Sprintf(fmt.Sprintf("board stream started at %s/api/board/game/stream/%s", hostUrl, game))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Accept", "application/x-ndjson")
	if err != nil {
		//ErrorMessage <- fmt.Errorf("stream event get request failed: %v", err)
		return fmt.Errorf("stream event get request failed: %v", err)
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		return fmt.Errorf("%v: %v", resp.StatusCode, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {

		return fmt.Errorf("bad response: %v", resp.StatusCode)
	}

	d := json.NewDecoder(resp.Body)
	//loop_i := 0
	for {
		select {
		// case <-quit_stream:
		// 	return nil
		default:
			//fmt.Printf("enter loop %v\n", loop_i)
			var responseData map[string]interface{}
			err := d.Decode(&responseData)
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)

					//fmt.Printf("decode error, not EOF\n")
				} else {

					//fmt.Printf("decode error, EOF\n")
					time.Sleep(1)
					close(event_chan)
					return nil
				}
			}

			if !isNil(responseData["type"]) { //make sure response is valid
				streamEvent := responseData["type"].(string)
				var bevent BoardEvent
				// bevent.Type = streamEvent
				// bevent.Moves = ""
				switch streamEvent {
				case "gameFull":

					// bevent.Rated = responseData["rated"].(bool)
					// state := responseData["state"].(map[string]interface{})
					// bevent.Moves = state["moves"].(string)
					// bevent.Status = state["status"].(string)
					//board_state_sig <- true
					bevent = GameFull
					jsonStr, err := json.Marshal(responseData)
					if err != nil {
						log.Fatal(err)
					}
					if err := json.Unmarshal(jsonStr, &BoardFullGame); err != nil {
						log.Fatal(err)
					}

				case "gameState":

					// bevent.Moves = responseData["moves"].(string)
					// bevent.Status = responseData["status"].(string)
					//board_state_sig <- true
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

				// if bevent.Moves != "" {
				// 	move_seq <- bevent.Moves
				// }

				//signal what event just happened so LichessGameScreen knows what global to look at
				event_chan <- bevent

				//fmt.Printf("%v\n", StreamEventID)
				//return nil
			} else {
				return fmt.Errorf("invalid gamestate stream")
			}
		}
	}
}

//consumes board events after the board stream has been started for a game

// func BoardConsumer(event_chan <-chan BoardEvent) {
// 	for {
// 		select {
// 		case e := <-event_chan:
// 			//fmt.Printf("consumer: %v %v \n", e.Event, e.Id)
// 			BoardStreamArr = append([]BoardState{e}, BoardStreamArr...)
// 			WriteLocal(fmt.Sprintf("BoardEvent_%v", time.Now().Format("2006-01-02_15:04:05")),
// 				fmt.Sprintf("%v\n%v\n%v\n%v", e.Type, e.Moves, e.Status, e.Rated))
// 		}
// 	}
// }
