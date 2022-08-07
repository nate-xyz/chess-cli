package pkg

import (
	"fmt"

	"github.com/nate-xyz/chess-cli/api"
)

func (user *Login) GetLichessUserInfo() error {
	var err error
	if user.Token != "" {
		if user.Email == "" {
			user.Email, err = api.GetEmail()
			if err != nil {
				return err
			}
		}
		if user.Name == "" {
			user.Name, err = api.GetUsername()
			if err != nil {
				return err
			}
		}
		user.Friends, err = api.GetFriends()
		if err != nil {
			return err
		}
		if !api.StreamEventStarted {
			Ready <- struct{}{} //start event stream
		}
		return nil
	}
	return fmt.Errorf("no token")
}

func (user *Login) GetOngoing() error {
	var err error
	user.OngoingGames, err = api.GetOngoingGames()
	if err != nil {
		return err
	}
	return nil
}

func (user *Login) GetChallenges() error {
	var err error
	user.IncomingChallenges, user.OutgoingChallenges, err = api.GetChallenges()
	if err != nil {
		return err
	}
	return nil
}
