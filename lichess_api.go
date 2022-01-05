package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var UserEmail string
var Username string
var UserProfile map[string]interface{}
var UserFriends string
var allFriends []string
var FriendsMap map[string]bool

//var Challenge map[string]interface{}

type TimeInfo struct {
	Increment int    `json: "increment"`
	Limit     int    `json: "limit"`
	Show      string `json: "show"`
	Type      string `json: "type"`
}

type VariantInfo struct {
	Key   string `json: "key"`
	Name  string `json: "name"`
	Short string `json: "short"`
}

type ChallengerInfo struct {
	Id     string `json: "id"`
	Name   string `json: "name"`
	Rating int    `json: "rating"`
	Title  string `json: "title"`
}

type Perf_ struct {
	Icon string `json: "icon"`
	Name string `json: "name"`
}

type ChallengeInfo struct {
	Id          string         `json: "id"`
	URL         string         `json: "url"`
	Color       string         `json: "color"`
	Direction   string         `json: "direction"`
	TimeControl TimeInfo       `json: "timeControl"`
	Variant     VariantInfo    `json: "variant"`
	Challenger  ChallengerInfo `json: "challenger"`
	DestUser    ChallengerInfo `json: "destUser"`
	Perf        Perf_          `json: "perf"`
	Rated       bool           `json: "rated"`
	Speed       string         `json: "speed"`
	Status      string         `json: "status"`
}

var IncomingChallenges []ChallengeInfo
var OutgoingChallenges []ChallengeInfo

var JSONresult struct {
	In  []ChallengeInfo `json: "in"`
	Out []ChallengeInfo `json: "out"`
}

type CreateChallengeType struct {
	Type           int
	Username       string
	DestUser       string
	Variant        string
	VariantIndex   int
	TimeOption     int
	ClockLimit     string
	ClockIncrement string
	Days           string
	Rated          string
	RatedBool      bool
	Color          string
	ColorIndex     int
	MinTurn        float64
	OpenEnded      bool
}

var ChallengeId string

//create a challenge against a specific user or get the url
func CreateChallenge(challenge CreateChallengeType) error {

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
	//application/x-www-form-urlencoded

	// create the request and execute it
	req, err := http.NewRequest("POST", requestUrl, strings.NewReader(reqParam.Encode()))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))

	if err != nil {
		ChallengeId = fmt.Sprintf("%v", err)
		return err
	}

	//do http request. must be done in this fashion so we can add the auth bear token headers above
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		ChallengeId = fmt.Sprintf("%v", err)
		return err
	}

	//read resp body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		err := fmt.Errorf("%v: %v", err, string(body))
		log.Fatalln(err)
	}
	defer res.Body.Close()

	if res.StatusCode == 400 {
		err := fmt.Errorf("Challenge creation failed: %v", err)
		return err
	}
	if res.StatusCode == 200 {
		return nil
	}
	return err
}

func StreamEvent() error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/stream/event", hostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	req.Header.Add("Content-Type", "application/x-ndjson")
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
		return fmt.Errorf(string(body))
	} else if err != nil {
		return err
	}

	// unmarshal the json into a string map
	//var responseData map[string]interface{}
	//err = json.Unmarshal(body, &responseData)
	allFriends = nil
	d := json.NewDecoder(strings.NewReader(string(body)))

	for {
		select {
		case <-quit_stream:
			return nil
		default:
			var responseData map[string]interface{}
			err := d.Decode(&responseData)
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
				}
				break
			}

			if !isNil(responseData["type"]) { // retrieve the username out of the map
				streamEvent = responseData["type"].(string)
			} else {
				return fmt.Errorf("username response interface is nil")
			}
		}
	}
	return nil
}

var streamEvent string

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
		//return fmt.Errorf(string(body))
		return fmt.Errorf("%v bad response", err)
	} else if err != nil {
		return err
	}
	// unmarshal the json into a string map
	//var responseData map[string]interface{}
	//var challengeList []interface{}
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
	req.Header.Add("Content-Type", "application/x-ndjson")
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
		return fmt.Errorf(string(body))
	} else if err != nil {
		return err
	}

	// unmarshal the json into a string map
	//var responseData map[string]interface{}
	//err = json.Unmarshal(body, &responseData)
	allFriends = nil
	d := json.NewDecoder(strings.NewReader(string(body)))
	for {
		var responseData map[string]interface{}
		err := d.Decode(&responseData)
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		//fmt.Printf("%v\n", responseData)

		var FriendsString string
		if !isNil(responseData["username"]) { // retrieve the username out of the map
			FriendsString = responseData["username"].(string)
			allFriends = append(allFriends, FriendsString)
			//fmt.Printf("%v\n", FriendsString)
			return nil
		} else {
			return fmt.Errorf("username response interface is nil")
		}

	}
	return nil

}
