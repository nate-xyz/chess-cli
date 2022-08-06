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

	d := json.NewDecoder(resp.Body)

	for {
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

				// jsonStr, err := json.Marshal(responseData)
				// if err != nil {
				// 	log.Fatal(err)
				// }
				// if err := json.Unmarshal(jsonStr, &CurrentStreamEventGame); err != nil {
				// 	log.Fatal(err)
				// }

			case "challenge", "challengeCanceled", "challengeDeclined": //challenge type stream event
				challenge := responseData["challenge"].(map[string]interface{})
				streamEventID = challenge["id"].(string)

				// jsonStr, err := json.Marshal(responseData)
				// if err != nil {
				// 	log.Fatal(err)
				// }
				// if err := json.Unmarshal(jsonStr, &CurrentStreamEventChallenge); err != nil {
				// 	log.Fatal(err)
				// }

			}
			EventChannel <- StreamEventType{streamEvent, streamEventID, streamSource}
		} else {
			Online = false
			return fmt.Errorf("invalid stream event")
		}
	}
}

//https://lichess.org/api/board/game/stream/{gameId}
//after initializing a game, this function streams the the state of the board of the game
func StreamBoardState(EventChannel chan<- BoardEvent, StreamError chan<- error, game string) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/board/game/stream/%s", HostUrl, game), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Content-Type", "application/x-ndjson")
	if err != nil {
		StreamError <- fmt.Errorf("stream event get request failed: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		StreamError <- fmt.Errorf("%v: %v", resp.Status, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		StreamError <- fmt.Errorf("%v", resp.Status)
	}

	d := json.NewDecoder(resp.Body)

	for {
		var responseData map[string]interface{}
		err := d.Decode(&responseData)

		if err != nil {
			if err != io.EOF {
				StreamError <- err
				return
			} else {
				EventChannel <- BoardEvent{Type: EOF}
				return
			}
		}

		if !isNil(responseData["type"]) { //make sure response is valid
			var bEvent BoardEvent
			streamEvent := responseData["type"].(string)

			switch streamEvent {

			case "gameFull":
				bEvent.Type = GameFull
				jsonStr, err := json.Marshal(responseData)
				if err != nil {
					log.Fatal(err)
				}
				if err := json.Unmarshal(jsonStr, &bEvent.Full); err != nil {
					log.Fatal(err)
				}

			case "gameState":
				bEvent.Type = GameState
				jsonStr, err := json.Marshal(responseData)
				if err != nil {
					log.Fatal(err)
				}
				if err := json.Unmarshal(jsonStr, &bEvent.State); err != nil {
					log.Fatal(err)
				}

			case "chatLine":
				bEvent.Type = ChatLine

			case "chatLineSpectator":
				bEvent.Type = ChatLineSpectator

			case "gameStateResign":
				bEvent.Type = GameStateResign
				jsonStr, err := json.Marshal(responseData)
				if err != nil {
					log.Fatal(err)
				}
				if err := json.Unmarshal(jsonStr, &bEvent.Resign); err != nil {
					log.Fatal(err)
				}

			}
			EventChannel <- bEvent //signal what event just happened
		} else {
			StreamError <- fmt.Errorf("invalid ndjson")
			return
		}
	}
}
