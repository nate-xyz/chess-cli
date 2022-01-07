package lichess

func containedInOngoingGames(a []OngoingGameInfo, gameid string) bool {
	for _, g := range a {
		if g.GameID == gameid {
			return true
		}
	}
	return false
}

func containedInEventStream(a []StreamEventType, gameid string) (string, bool) {
	for _, e := range a {
		if e.Id == gameid {
			return e.Event, true
		}
	}
	return "", false
}

func getEvents(a []StreamEventType, gameid string) ([]StreamEventType, bool) {
	n := make([]StreamEventType, 0)
	for _, e := range a {
		if e.Id == gameid {
			n = append(n, e)
		}
	}
	if len(n) > 0 {
		return n, true
	}
	return n, false
}
