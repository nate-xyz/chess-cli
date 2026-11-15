package pkg

import (
	"fmt"

	"github.com/nate-xyz/chess-cli/api"
)

// Display preferences from Options.

// clockBarsOn: clocks as a bar beside the board (else plain text in the sidebar).
func clockBarsOn() bool { return !api.UserInfo.ClockBarOff }

// zenMode hides distractions in games, like Lichess's zen mode: names,
// ratings, chat and game info. The board, clocks and moves stay.
func zenMode() bool { return api.UserInfo.Zen }

func setClockBars(on bool) {
	_ = api.UpdateConfig(func(c *api.UserConfig) { c.ClockBarOff = !on })
	applyLayoutPrefs()
}

func setZen(on bool) {
	_ = api.UpdateConfig(func(c *api.UserConfig) { c.Zen = on })
	applyLayoutPrefs()
}

// applyLayoutPrefs shows or hides the parts of the game screens these
// settings affect. Called when a setting changes and when a screen opens.
func applyLayoutPrefs() {
	if og := Root.ongame; og != nil && og.Side != nil {
		chat, input := 2, 1
		if zenMode() {
			chat, input = 0, 0
		}
		og.Side.ResizeItem(og.Chat, 0, chat)
		og.Side.ResizeItem(og.ChatInput, input, 0)
		timer := 0
		if !clockBarsOn() {
			timer = 1
			og.Board.HideClockBars()
		}
		og.Side.ResizeItem(og.OppTimer, timer, 0)
		og.Side.ResizeItem(og.UserTimer, timer, 0)
		if Root.gameState != nil && og.Full.ID != "" { // a game is loaded
			og.UpdateUserInfo()
			og.UpdateStatus()
		}
	}
	if ws := Root.watch; ws != nil && ws.Side != nil {
		timer := 0
		if !clockBarsOn() {
			timer = 1
		}
		ws.Side.ResizeItem(ws.TopClock, timer, 0)
		ws.Side.ResizeItem(ws.BotClock, timer, 0)
		ws.updateInfo()
		ws.updateClocks()
	}
}

// timeControlText describes a time control: "10+5", "½+0",
// "Correspondence · 3 days", "∞ Unlimited". Times in seconds.
func timeControlText(initial, increment int, speed string, days int) string {
	switch {
	case speed == "correspondence" || days > 0:
		if days == 1 {
			return "Correspondence · 1 day"
		}
		if days > 1 {
			return fmt.Sprintf("Correspondence · %d days", days)
		}
		return "Correspondence"
	case speed == "unlimited" || initial == 0 && increment == 0:
		return "∞ Unlimited"
	}
	minutes := map[int]string{15: "¼", 30: "½", 45: "¾", 90: "1½"}[initial]
	if minutes == "" {
		minutes = fmt.Sprint(initial / 60)
	}
	return fmt.Sprintf("%s+%d", minutes, increment)
}
