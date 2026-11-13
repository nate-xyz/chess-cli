package pkg

import (
	"sync"

	"github.com/nate-xyz/chess-cli/api"
)

// Account events (game starts, challenges) arrive on one stream. StreamConsumer
// is its only reader and fans each event out to whoever has subscribed, so
// screens that wait for a game (the loader) don't race each other for events.
var (
	eventsMu  sync.Mutex
	eventSubs = map[int]chan api.StreamEventType{}
	nextSubID int
)

// subscribeEvents returns a channel of account events and a function to stop
// receiving them.
func subscribeEvents() (<-chan api.StreamEventType, func()) {
	eventsMu.Lock()
	defer eventsMu.Unlock()
	id := nextSubID
	nextSubID++
	ch := make(chan api.StreamEventType, 16)
	eventSubs[id] = ch
	return ch, func() {
		eventsMu.Lock()
		defer eventsMu.Unlock()
		delete(eventSubs, id)
	}
}

// StreamConsumer reads account events and distributes them to subscribers.
func StreamConsumer(events <-chan api.StreamEventType) {
	for e := range events {
		Root.App.QueueUpdateDraw(func() {
			if name, _ := Root.nav.GetFrontPanel(); name == "lichesswelcome" {
				Root.wonline.UpdateTitle("")
			}
		})
		eventsMu.Lock()
		for _, ch := range eventSubs {
			select {
			case ch <- e:
			default: // a slow subscriber shouldn't stall the stream
			}
		}
		eventsMu.Unlock()
	}
}
