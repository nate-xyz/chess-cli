package pkg

import (
	"fmt"

	"github.com/nate-xyz/chess-cli/api"
)

func (user *Login) GetLichessUserInfo() error {
	var err error
	if user.Token != "" {
		if user.Name == "" {
			user.Name, err = api.GetUsername()
			if err != nil {
				return err
			}
		}
		user.Friends, _ = api.GetFriends() // only used for suggestions; not worth failing login over
		if StreamChannel != nil {
			api.StartEventStream(StreamChannel)
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
