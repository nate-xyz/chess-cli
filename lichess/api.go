package lichess

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	. "github.com/nate-xyz/chess-cli/shared"
	//	. "github.com/nate-xyz/chess-cli/shared"
)

func StreamBoardState(event_chan chan<- BoardState, board_state_sig chan<- bool, game string, ErrorMessage chan<- error) error {
	//https://lichess.org/api/board/game/stream/{gameId}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/board/game/stream/%s", hostUrl, game), nil)
	NotiMessage <- fmt.Sprintf(fmt.Sprintf("board stream started at %s/api/board/game/stream/%s", hostUrl, game))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Accept", "application/x-ndjson")
	if err != nil {
		ErrorMessage <- fmt.Errorf("stream event get request failed: %v", err)
		return fmt.Errorf("stream event get request failed: %v", err)
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		ErrorMessage <- fmt.Errorf("%v: %v", resp.StatusCode, err)
		return fmt.Errorf("%v: %v", resp.StatusCode, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		ErrorMessage <- fmt.Errorf("bad response: %v", resp.StatusCode)
		return fmt.Errorf("bad response: %v", resp.StatusCode)
	} else {
		NotiMessage <- fmt.Sprintf("board stream good response: %v", resp.StatusCode)
		//ErrorMessage <- fmt.Errorf("good response: %v", resp.StatusCode)
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
					ErrorMessage <- fmt.Errorf("decode error, not EOF: %v", err)
					//fmt.Printf("decode error, not EOF\n")

				}
				ErrorMessage <- fmt.Errorf("decode error, EOF: %v", err)
				//fmt.Printf("decode error, EOF\n")
				continue
			}

			// Type   string
			// Moves  string
			// Status string
			// Rated  string

			if !isNil(responseData["type"]) { // retrieve the username out of the map
				streamEvent = responseData["type"].(string)
				var bevent BoardState
				bevent.Type = streamEvent
				switch streamEvent {
				case "gameFull":
					NotiMessage <- fmt.Sprintf("got a gamestate")
					bevent.Rated = responseData["rated"].(bool)
					state := responseData["state"].(map[string]interface{})
					bevent.Moves = state["moves"].(string)
					bevent.Status = state["status"].(string)
					board_state_sig <- true
				case "gameState":
					NotiMessage <- fmt.Sprintf("got a gamestate")
					bevent.Moves = responseData["moves"].(string)
					bevent.Status = responseData["status"].(string)
					board_state_sig <- true
				case "chatLine":
					//TODO

				}

				event_chan <- bevent

				//fmt.Printf("%v\n", StreamEventID)
				//return nil
			} else {
				ErrorMessage <- fmt.Errorf("no type in stream event")
				return fmt.Errorf("no type in stream event")
			}
		}
	}
}

func BoardConsumer(event_chan <-chan BoardState, noti chan<- string) {
	for {
		select {
		case e := <-event_chan:
			//fmt.Printf("consumer: %v %v \n", e.Event, e.Id)
			BoardStreamArr = append([]BoardState{e}, BoardStreamArr...)
			noti <- fmt.Sprintf("event %v", e.Type)
		}
	}
}

func StreamEvent(event_chan chan<- StreamEventType, got_token chan struct{}) error {
	<-got_token
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/stream/event", hostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Accept", "application/x-ndjson")
	if err != nil {
		return fmt.Errorf("stream event get request failed: %v", err)
	}
	NotiMessage <- fmt.Sprintf("event stream started at %s/api/stream/event", hostUrl)
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

				}
				//fmt.Printf("decode error, EOF\n")
				continue
			}
			if !isNil(responseData["type"]) { // retrieve the username out of the map
				streamEvent = responseData["type"].(string)
				//fmt.Printf("received stream event! -> %v: ", streamEvent)
				var streamEventID string
				switch streamEvent {
				case "gameStart", "gameFinish":
					chal := responseData["game"].(map[string]interface{})
					streamEventID = chal["id"].(string)
				case "challenge", "challengeCanceled", "challengeDeclined":
					chal := responseData["challenge"].(map[string]interface{})
					streamEventID = chal["id"].(string)

				}
				event_chan <- StreamEventType{streamEvent, streamEventID}

				//ErrorMessage <- fmt.Errorf("%v, %v", streamEvent, streamEventID)
				//fmt.Printf("%v\n", StreamEventID)
				//return nil
			} else {
				return fmt.Errorf("no type in stream event")
			}
		}
		NotiMessage <- "stream still open"
		//fmt.Printf("end loop %v\n", loop_i)
		//loop_i++
	}
}

func StreamConsumer(event_chan <-chan StreamEventType, noti chan<- string) {
	for {
		select {
		case e := <-event_chan:
			//fmt.Printf("consumer: %v %v \n", e.Event, e.Id)
			EventStreamArr = append([]StreamEventType{e}, EventStreamArr...)
			StreamChannelForWaiter <- e
			if e.Event != "gameStart" {
				noti <- fmt.Sprintf("event %v:%v", e.Event, e.Id)
			} else {
				if !containedInOngoingGames(OngoingGames, e.Id) {
					noti <- fmt.Sprintf("new game! %v:%v", e.Event, e.Id)
				}
			}
			//noti <- fmt.Sprintf("event %v:%v", e.Event, e.Id)
			//CheckGlobal = e.Id

			//TODO: add to global data structure storing all events
		}
	}
}

//create a challenge against a specific user or get the url
func CreateChallenge(challenge CreateChallengeType) (error, string) {
	requestUrl := fmt.Sprintf("%s/api/challenge/%s", hostUrl, challenge.DestUser)
	var reqParam url.Values
	switch challenge.TimeOption {
	case 0: //realtime
		reqParam = url.Values{
			"rated":           {challenge.Rated},
			"clock.limit":     {challenge.ClockLimit},
			"clock.increment": {challenge.ClockIncrement},
			"color":           {challenge.Color},
			"variant":         {challenge.Variant},
			"keepAliveStream": {"true"},
		}
	case 1: //corresondesnce
		reqParam = url.Values{
			"rated":           {challenge.Rated},
			"days":            {challenge.Days},
			"color":           {challenge.Color},
			"variant":         {challenge.Variant},
			"keepAliveStream": {"true"},
		}
	case 2: //unlimited
		reqParam = url.Values{
			"rated":           {challenge.Rated},
			"color":           {challenge.Color},
			"variant":         {challenge.Variant},
			"keepAliveStream": {"true"},
		}

	}
	NotiMessage <- fmt.Sprintf("%s", reqParam)
	//application/x-www-form-urlencoded

	// create the request and execute it
	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(reqParam.Encode()))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	NotiMessage <- fmt.Sprintf("POST request at %s", requestUrl)

	if err != nil {
		return err, ""
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err, ""
	}

	//read resp body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		err := fmt.Errorf("%v", err)
		//log.Fatalln(err)
		return err, ""
	}
	defer res.Body.Close()

	//fmt.Printf("%v", res.StatusCode)
	//fmt.Printf(string(body))

	if res.StatusCode == 400 {
		err := fmt.Errorf("Challenge creation failed: %v", err)
		return err, ""
	}
	if res.StatusCode != 200 {
		err := fmt.Errorf("not 200 response: %v", err)
		return err, ""
	} else {
		NotiMessage <- fmt.Sprintf("challenge good response: %v", res.StatusCode)
	}
	// unmarshal the json into a string map
	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return err, ""
	}

	// retrieve the access token out of the map, and return to caller
	if !isNil(responseData["challenge"]) {
		NotiMessage <- fmt.Sprintf("posted challenge\n")
		chal := responseData["challenge"].(map[string]interface{})
		ChallengeId = chal["id"].(string)
		return nil, ChallengeId
	}
	// d := json.NewDecoder(strings.NewReader(string(body)))
	// for {
	// 	select {
	// 	// case <-quit_stream:
	// 	// 	return nil
	// 	default:
	// 		var responseData map[string]map[interface]interface{}

	// 		err := d.Decode(&responseData)
	// 		if err != nil {
	// 			if err != io.EOF {
	// 				log.Fatal(err)
	// 				return err
	// 			}
	// 			continue
	// 		}
	// 		if !isNil(responseData["challenge"]) { // retrieve the username out of the map
	// 			streamEvent = responseData["challenge"]["id"].(string)
	// 			fmt.Printf(streamEvent)
	// 			return nil
	// 		}
	// 		//fmt.Printf("waiting 2")
	// 		//return fmt.Errorf("username response interface is nil")
	// 	}
	// }
	return fmt.Errorf("username response interface is nil"), ""
}

func GetOngoingGames() error {

	// var reqParam url.Values
	// reqParam = url.Values{
	// 	"nb": {"50"},
	// }
	//p := strings.NewReader(reqParam.Encode())
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/account/playing", hostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	//req.Header.Add("Content-Type", "application/json")

	if err != nil {
		return err
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		err := fmt.Errorf("response: %v", resp.StatusCode)
		return err
	}
	// unmarshal the json into a string map
	var JSONresult struct {
		NowPlaying []OngoingGameInfo `json: "nowPlaying"`
	}
	err = json.Unmarshal(body, &JSONresult)
	if err != nil {
		return err
	}

	// retrieve the access token out of the map, and return to caller
	if !isNil(JSONresult.NowPlaying) {
		OngoingGames = JSONresult.NowPlaying
		return nil
	}
	return fmt.Errorf("response interface is nil")
}

func GetChallenges() error {

	//http GET returns array of objects(ChallengeJson) in and out
	//ChallengesArray := make([]string, 0)
	//var IncomingChallenges string = ""
	//var OutgoingChallenges string = ""

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/challenge", hostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	//req.Header.Add("Content-Type", "application/json")

	if err != nil {
		return err
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	//read resp body
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("%v bad response", err)
	} else if err != nil {
		return err
	}
	// unmarshal the json into a string map
	//var responseData map[string]interface{}
	//var challengeList []interface{}
	//var JSONresult map[string]interface{}
	var JSONresult struct {
		In  []ChallengeInfo `json: "in"`
		Out []ChallengeInfo `json: "out"`
	}

	err = json.Unmarshal(body, &JSONresult)
	if err != nil {
		return err
	}
	//fmt.Println(string(body))
	//res2B, _ := json.Marshal(JSONresult)
	//fmt.Println(string(res2B))
	// retrieve the access token out of the map, and return to caller

	if !isNil(JSONresult.In) {
		IncomingChallenges = JSONresult.In

		// challengeList = responseData["in"].([]interface{})
		// for _, challenge := range challengeList {
		// 	challenge_ := challenge.(ChallengeInfo)
		// 	fmt.Printf("%v", challenge_.UserID)
		// }
		//IncomingChallenges = responseData["out"].([]ChallengeInfo)
		//json.Unmarshal([]byte(responseData["in"]), &IncomingChallenges)
		//ChallengesArray = append(ChallengesArray, IncomingChallenges)
	} else {
		return fmt.Errorf("response interface is nil")
	}
	if !isNil(JSONresult.Out) {
		OutgoingChallenges = JSONresult.Out
		// challengeList := JSONresult["out"].([]interface{})
		// for _, challenge := range challengeList {
		// 	//challenge_ := challenge.(ChallengeInfo)
		// 	fmt.Printf("%v", challenge.(map[string]interface{})["id"])
		// }
		//OutgoingChallenges = responseData["out"].([]ChallengeInfo)
		//ChallengesArray = append(ChallengesArray, OutgoingChallenges)
		return nil
	}
	return fmt.Errorf("response interface is nil")
}

//get user email
func GetEmail() error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/account/email", hostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	if err != nil {
		return err
	}
	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf(string(body))
	}
	// unmarshal the json into a string map
	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return err
	}

	// retrieve the access token out of the map, and return to caller
	if !isNil(responseData["email"]) {
		UserEmail = responseData["email"].(string)
		return nil
	}
	return fmt.Errorf("response interface is nil")

}

//get username from profile json
func GetUsername() error {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/account", hostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	if err != nil {
		return err
	}
	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf(string(body))
	}
	// unmarshal the json into a string map
	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return err
	}

	// retrieve the username out of the map, and return to caller
	if !isNil(responseData["username"]) {
		Username = responseData["username"].(string)
		return nil
	}
	return fmt.Errorf("response interface is nil")

}

//get full user profile json
func GetProfile() error {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/account", hostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	if err != nil {
		return err
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf(string(body))
	} else if err != nil {
		return err
	}
	// unmarshal the json into a string map
	err = json.Unmarshal(body, &UserProfile)
	if err != nil {
		return err
	}
	return nil
}

//application/x-ndjson
//list of friends(and their online/offline status), to be displayed on challenge screen
func GetFriends() error {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/rel/following", hostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Accept", "application/x-ndjson")
	if err != nil {
		return err
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

				}
				//fmt.Printf("decode error, EOF\n")
				return nil
			}

			if !isNil(responseData["username"]) { // retrieve the username out of the map
				FriendsString := responseData["username"].(string)
				allFriends = append(allFriends, FriendsString)
				//fmt.Printf("%v\n", FriendsString)
			} else {
				return fmt.Errorf("no type in stream event")
			}
		}
	}

}
