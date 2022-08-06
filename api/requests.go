package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

//create a challenge against a specific user or get the url
func MakeMove(gameid string, move string) error {
	requestUrl := fmt.Sprintf("%s/api/board/game/%s/move/%s", HostUrl, gameid, move)

	// create the request and execute it
	req, err := http.NewRequest("POST", requestUrl, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	if err != nil {
		return err
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		//read resp body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			err := fmt.Errorf("%v", err)
			return err
		}
		defer res.Body.Close()

		err = fmt.Errorf("%s\n%s", res.Status, string(body))
		return err
	}
	return nil
}

//create a challenge against a specific user or get the url (POST)
func CreateChallenge(challenge CreateChallengeType) (error, string) {
	requestUrl := fmt.Sprintf("%s/api/challenge/%s", HostUrl, challenge.DestUser)
	var reqParam url.Values
	switch challenge.TimeOption {
	case 0: //realtime
		reqParam = url.Values{
			"rated":           {challenge.Rated},
			"clock.limit":     {challenge.ClockLimit},
			"clock.increment": {challenge.ClockIncrement},
			"color":           {challenge.Color}, //enum: 0, 1, or 2
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

	//application/x-www-form-urlencoded

	// create the request and add headers
	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(reqParam.Encode()))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

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
	body, err := io.ReadAll(res.Body) //TODO: display body error on loading screen
	if err != nil {
		err := fmt.Errorf("%v", err)
		return err, ""
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("%s", res.Status)
		return err, ""
	}

	d := json.NewDecoder(strings.NewReader(string(body)))
	for {
		select {
		// case <-quit_stream:
		// 	return nil
		default:
			var responseData map[string]interface{}
			err = d.Decode(&responseData)
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
					return err, ""
				}
				continue
			}
			if !isNil(responseData["challenge"]) { // retrieve the username out of the map
				challenge := responseData["challenge"].(map[string]interface{})
				ChallengeId = challenge["id"].(string)
				return nil, ChallengeId
			}
			//fmt.Printf("waiting 2")
			//return fmt.Errorf("username response interface is nil")
		}
	}
	//return fmt.Errorf("username response interface is nil"), ""
}

func CreateAiChallenge(challenge CreateChallengeType) (error, string) {
	requestUrl := fmt.Sprintf("%s/api/challenge/ai", HostUrl)
	var reqParam url.Values
	switch challenge.TimeOption {
	case 0: //realtime
		reqParam = url.Values{
			"level":           {challenge.Level},
			"clock.limit":     {challenge.ClockLimit},
			"clock.increment": {challenge.ClockIncrement},
			"color":           {challenge.Color}, //enum: 0, 1, or 2
			"variant":         {challenge.Variant},
			"keepAliveStream": {"true"},
		}
	case 1: //corresondesnce
		reqParam = url.Values{
			"level":           {challenge.Level},
			"days":            {challenge.Days},
			"color":           {challenge.Color},
			"variant":         {challenge.Variant},
			"keepAliveStream": {"true"},
		}
	case 2: //unlimited
		reqParam = url.Values{
			"level":           {challenge.Level},
			"color":           {challenge.Color},
			"variant":         {challenge.Variant},
			"keepAliveStream": {"true"},
		}
	}

	//application/x-www-form-urlencoded

	// create the request and add headers
	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(reqParam.Encode()))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")

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
	body, err := io.ReadAll(res.Body) //TODO: display body error on loading screen
	if err != nil {
		err := fmt.Errorf("%v", err)
		//log.Fatalln(err)
		return err, ""
	}
	defer res.Body.Close()

	//fmt.Printf("%v", res.StatusCode)
	//fmt.Printf(string(body))

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		err := fmt.Errorf("%s", res.Status)
		return err, ""
	}

	d := json.NewDecoder(strings.NewReader(string(body)))
	for {
		select {
		// case <-quit_stream:
		// 	return nil
		default:
			var responseData map[string]interface{}
			err = d.Decode(&responseData)
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
					return err, ""
				}
				continue
			}
			if !isNil(responseData["id"]) { // retrieve the username out of the map
				ChallengeId := responseData["id"].(string)
				return nil, ChallengeId
			}
			//fmt.Printf("waiting 2")
			//return fmt.Errorf("username response interface is nil")
		}
	}
	//return fmt.Errorf("username response interface is nil"), ""
}

func CreateSeek(challenge CreateChallengeType) error {
	requestUrl := fmt.Sprintf("%s/api/board/seek", HostUrl)
	var reqParam url.Values
	switch challenge.TimeOption {
	case 0: //realtime
		reqParam = url.Values{
			"rated":           {challenge.Rated},
			"clock.limit":     {challenge.ClockLimit},
			"clock.increment": {challenge.ClockIncrement},
			"color":           {challenge.Color}, //enum: 0, 1, or 2
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

	//application/x-www-form-urlencoded

	// create the request and add headers
	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(reqParam.Encode()))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Content-type", "application/x-www-form-urlencoded")

	if err != nil {
		return err
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	//read resp body
	body, err := io.ReadAll(res.Body) //TODO: display body error on loading screen
	if err != nil {
		err := fmt.Errorf("%v", err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("%s\n%s", res.Status, string(body))
		return err
	} else {
		return nil
	}
}

func GetOngoingGames() error {
	requestUrl := fmt.Sprintf("%s/api/account/playing", HostUrl)
	reqParam := url.Values{"nb": {"11-50"}}
	req, err := http.NewRequest("GET", requestUrl, strings.NewReader(reqParam.Encode()))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Content-Type", "application/json")

	if err != nil {
		return err
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("bad response: %v", resp.StatusCode)
		return err
	}
	d := json.NewDecoder(resp.Body)
	OngoingGames = nil
	for {
		var responseData map[string]interface{}
		err := d.Decode(&responseData)
		if err != nil {
			if err != io.EOF {
				return err
			} else {
				return nil
			}
		}

		if !isNil(responseData["nowPlaying"]) { //make sure response is valid
			jsonStr, err := json.Marshal(responseData["nowPlaying"])
			if err != nil {
				log.Fatal(err)
			}
			var OngoingGameList []OngoingGameInfo
			if err := json.Unmarshal(jsonStr, &OngoingGameList); err != nil {
				log.Fatal(err)
			}
			OngoingGames = append(OngoingGames, OngoingGameList...)
		}
	}
}

func GetChallenges() error {
	requestUrl := fmt.Sprintf("%s/api/challenge", HostUrl)
	req, err := http.NewRequest("GET", requestUrl, nil)
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

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad response: %v ", resp.Status)
		return err
	}
	d := json.NewDecoder(resp.Body)
	IncomingChallenges = nil
	OutgoingChallenges = nil

	for {
		var responseData map[string]interface{}
		err := d.Decode(&responseData)
		if err != nil {
			if err != io.EOF {
				return err
			} else {
				return nil
			}
		}

		if !isNil(responseData["in"]) { //make sure response is valid
			jsonStr, err := json.Marshal(responseData["in"])
			if err != nil {
				log.Fatal(err)
			}
			var In []ChallengeInfo
			if err := json.Unmarshal(jsonStr, &In); err != nil {
				log.Fatal(err)
			}
			IncomingChallenges = append(IncomingChallenges, In...)
		}

		if !isNil(responseData["out"]) { //make sure response is valid
			jsonStr, err := json.Marshal(responseData["out"])
			if err != nil {
				log.Fatal(err)
			}
			var Out []ChallengeInfo
			if err := json.Unmarshal(jsonStr, &Out); err != nil {
				log.Fatal(err)
			}
			OutgoingChallenges = append(OutgoingChallenges, Out...)
		}

	}
}

//get user email
func GetEmail() error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/account/email", HostUrl), nil)
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
	if resp.StatusCode != http.StatusOK {
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
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/account", HostUrl), nil)
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
	if resp.StatusCode != http.StatusOK {
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
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/account", HostUrl), nil)
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
	if resp.StatusCode != http.StatusOK {
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
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/rel/following", HostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Content-Type", "application/x-ndjson")
	if err != nil {
		return err
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GET FRIENDS %v: %v", resp.StatusCode, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET FRIENDS bad response: %v", resp.StatusCode)
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
				AllFriends = append(AllFriends, FriendsString)
				//fmt.Printf("%v\n", FriendsString)
			} else {
				return fmt.Errorf("no type in stream event")
			}
		}
	}

}

func AcceptChallenge(gameid string) error {
	requestUrl := fmt.Sprintf("%s/api/challenge/%s/accept", HostUrl, gameid)

	// create the request and execute it
	req, err := http.NewRequest("POST", requestUrl, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	if err != nil {
		return err
	}

	//do http request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("reponse %v", res.Status)
		return err
	}
	return nil
}

func AbortGame(gameid string) error {
	requestUrl := fmt.Sprintf("%s/api/board/game/%s/abort", HostUrl, gameid)

	// create the request and execute it
	req, err := http.NewRequest("POST", requestUrl, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	if err != nil {
		return err
	}

	//do http request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("%v", res.Status)
		return err
	}
	return nil
}

func ResignGame(gameid string) error {
	requestUrl := fmt.Sprintf("%s/api/board/game/%s/resign", HostUrl, gameid)

	// create the request and execute it
	req, err := http.NewRequest("POST", requestUrl, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	if err != nil {
		return err
	}

	//do http request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("%v", res.Status)
		return err
	}
	return nil
}

func HandleDraw(gameid string, accept bool) error {
	var acceptStr string
	if accept {
		acceptStr = "yes"
	} else {
		acceptStr = "no"
	}
	requestUrl := fmt.Sprintf("%s/api/board/game/%s/draw/%s", HostUrl, gameid, acceptStr)

	// create the request and execute it
	req, err := http.NewRequest("POST", requestUrl, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	if err != nil {
		return err
	}

	//do http request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("%v", res.Status)
		return err
	}
	return nil

}

func HandleTakeback(gameid string, accept bool) error {
	var acceptStr string
	if accept {
		acceptStr = "yes"
	} else {
		acceptStr = "no"
	}
	requestUrl := fmt.Sprintf("%s/api/board/game/%s/takeback/%s", HostUrl, gameid, acceptStr)

	// create the request and execute it
	req, err := http.NewRequest("POST", requestUrl, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	if err != nil {
		return err
	}

	//do http request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("%v", res.Status)
		return err
	}
	return nil

}
