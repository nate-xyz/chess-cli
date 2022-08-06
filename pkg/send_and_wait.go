package pkg

import (
	"fmt"
	"log"
	"time"

	"github.com/nate-xyz/chess-cli/api"
)

func SendChallengeRequest(gchan chan<- string, errchan chan<- error) {
	if api.UserInfo.ApiToken != "" {
		var err error
		var id string
		switch CurrentChallenge.Type {
		case 0:
			err = api.CreateSeek(CurrentChallenge)
			if err == nil {
				gchan <- "Seek created. Searching for opponent..."
				return
			}
		case 1: //challenge friend
			err, id = api.CreateChallenge(CurrentChallenge)
		case 2: //lichess ai
			err, id = api.CreateAiChallenge(CurrentChallenge)
		}
		if err != nil {
			errchan <- err
		} else {
			gchan <- id
		}
	} else {
		log.Fatal(fmt.Errorf("tried to send api request with no api token"))
	}
}

func WaitForLichessGameResponse() {
	var localid string
	getgameid := make(chan string, 1)
	ticker := time.NewTicker(time.Millisecond * 250)
	var icon_index int = 0
	RequestError := make(chan error, 1)
	var RandomSeek bool = CurrentChallenge.Type == 0
	go SendChallengeRequest(getgameid, RequestError) //send the request
	load_msg := "Requesting new game from Lichess, please wait."
	Root.loader.DrawMessage(load_msg)
	for {
		select {
		case e := <-RequestError:
			load_msg = fmt.Sprintf("ERROR: %v", e)
			Root.loader.DrawMessage(load_msg)
			go func() {
				time.Sleep(time.Second * 5)
				Root.App.QueueUpdate(gotoLichessAfterLogin)
			}()
		case id := <-getgameid:
			if !RandomSeek {
				localid = id
			} else {
				load_msg = id
				Root.loader.DrawMessage(load_msg)
			}
		case e := <-StreamChannel: // receive event directly
			EventStreamArr = append([]api.StreamEventType{e}, EventStreamArr...)
			if !RandomSeek { //friend or AI challenge
				if e.GameID == localid {
					load_msg = "Game started, going to game screen. (FROM ARRAY)"
					Root.loader.DrawMessage(load_msg)
					currentGameID = e.GameID
					Root.App.QueueUpdate(startNewOnlineGame) //goto game
					return                                   //necessary for QueueUpdate
				}
			} else {
				if e.EventType == "gameStart" && e.Source != "friend" { //TODO: check to make sure match random seek request
					load_msg = "Found random opponent!"
					Root.loader.DrawMessage(load_msg)
					go func() {
						currentGameID = e.GameID
						time.Sleep(time.Second)
						Root.App.QueueUpdate(startNewOnlineGame)
					}()
				}
			}
		case <-ticker.C:
			if !RandomSeek {
				if localid != "" {
					s, b := containedInEventStream(EventStreamArr, localid)
					if b {
						switch s {
						case "challenge":
							load_msg = fmt.Sprintf("Waiting for %v to accept the challenge %v/%v.", CurrentChallenge.DestUser, api.HostUrl, localid)
							Root.loader.DrawMessage(load_msg)
						case "gameFinish":
							return
						case "challengeCanceled", "challengeDeclined":
							return
						case "gameStart":
							load_msg = "Game started, going to game screen. (FROM STREAM)"
							Root.loader.DrawMessage(load_msg)
							currentGameID = localid
							Root.App.QueueUpdate(startNewOnlineGame) ///goto game
							return                                   //necessary for QueueUpdate
						}
					} else {
						load_msg = fmt.Sprintf("Waiting for %v to accept the challenge %v/%v.", CurrentChallenge.DestUser, api.HostUrl, localid)
						Root.loader.DrawMessage(load_msg)
					}
				}
			} else {
				e, b := EventContainedInEventStream(EventStreamArr, "gameStart") //TODO: check to make sure match random seek request
				if b && (e.Source != "friend") {
					load_msg = "Found random opponent!"
					Root.loader.DrawMessage(load_msg)
					go func() {
						currentGameID = e.GameID
						time.Sleep(time.Second)
						Root.App.QueueUpdate(startNewOnlineGame)
					}()
				}
			}
			icon_index = Root.loader.DrawIcon(icon_index)
		default:
			Root.loader.DrawMessage(load_msg)
		}
	}
}
