package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var UserEmail string
var Username string
var UserProfile map[string]interface{}
var UserFriends string
var allFriends []string
var FriendsMap map[string]bool

var Challenge map[string]interface{}

var IncomingChallenges []map[string]interface{}
var OutgoingChallenges []map[string]interface{}

func GetChallenges() error {

	//http GET returns array of objects(ChallengeJson) in and out
	//ChallengesArray := make([]string, 0)
	//var IncomingChallenges string = ""
	//var OutgoingChallenges string = ""

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/account/challenge", hostUrl), nil)
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

	//read resp body
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf(string(body))
	} else if err != nil {
		return err
	}
	// unmarshal the json into a string map
	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)

	// retrieve the access token out of the map, and return to caller
	if !isNil(responseData["in"]) {
		IncomingChallenges = responseData["in"].([]map[string]interface{})
		//json.Unmarshal([]byte(responseData["in"]), &IncomingChallenges)
		//ChallengesArray = append(ChallengesArray, IncomingChallenges)
	} else {
		return fmt.Errorf("response interface is nil")
	}
	if !isNil(responseData["out"]) {
		OutgoingChallenges = responseData["out"].([]map[string]interface{})
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
		} else {
			return fmt.Errorf("username response interface is nil")
		}

	}
	return nil

}
