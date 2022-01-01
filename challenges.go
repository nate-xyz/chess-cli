package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetChallenges() {

	//http GET returns array of objects(ChallengeJson) in and out
	ChallengesArray := make([]string, 0)
	var IncomingChallenges string = ""
	var OutgoingChallenges string = ""

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/account/challenge", hostUrl), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", UserInfo.ApiToken))
	if err != nil {
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	body, _ := io.ReadAll(resp.Body)

	// unmarshal the json into a string map
	var responseData map[string]interface{}
	err = json.Unmarshal(body, &responseData)

	// retrieve the access token out of the map, and return to caller
	if !isNil(responseData["in"]) {

		IncomingChallenges = responseData["in"].(string)
		ChallengesArray = append(ChallengesArray, IncomingChallenges)
	}

	if !isNil(responseData["out"]) {

		OutgoingChallenges = responseData["out"].(string)
		ChallengesArray = append(ChallengesArray, OutgoingChallenges)
	}
}
