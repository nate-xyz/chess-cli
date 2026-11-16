package pkg

import (
	"strings"
	"testing"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func TestTimeControlText(t *testing.T) {
	for _, c := range []struct {
		initial, inc int
		speed        string
		days         int
		want         string
	}{
		{600, 5, "rapid", 0, "10+5"},
		{180, 2, "blitz", 0, "3+2"},
		{30, 0, "ultraBullet", 0, "½+0"},
		{0, 0, "correspondence", 3, "Correspondence · 3 days"},
		{0, 0, "correspondence", 1, "Correspondence · 1 day"},
		{0, 0, "unlimited", 0, "∞ Unlimited"},
	} {
		if got := timeControlText(c.initial, c.inc, c.speed, c.days); got != c.want {
			t.Errorf("timeControlText(%d, %d, %q, %d) = %q, want %q", c.initial, c.inc, c.speed, c.days, got, c.want)
		}
	}
}

// Changing display options with no game loaded must be safe.
func TestApplyLayoutPrefsWithoutGame(t *testing.T) {
	startTestApp(t, 160, 50)
	onUI(func() {
		setZen(true)
		setClockBars(false)
		setZen(false)
		setClockBars(true)
	})
}

func TestRandomQuoteNeverEmpty(t *testing.T) {
	for i := 0; i < 500; i++ {
		if GetRandomQuote() == "" {
			t.Fatal("empty quote")
		}
	}
}

// Reset all settings: defaults everywhere, the login kept, and the
// first-run setup shown again.
func TestResetSettings(t *testing.T) {
	startTestApp(t, 160, 50)
	onUI(func() {
		_ = api.UpdateConfig(func(c *api.UserConfig) {
			c.ApiToken, c.Zen, c.Sound, c.MoveBoxAlways, c.Pieces, c.SoundAsked = "tok", true, SoundWood, true, PiecesBasic, true
		})
		resetSettings()
	})
	onUI(func() {
		defer func() {
			_ = api.UpdateConfig(func(c *api.UserConfig) { *c = api.UserConfig{Pieces: PiecesText, SoundAsked: true} })
		}()
		c := api.UserInfo
		if c.ApiToken != "tok" {
			t.Error("reset lost the login")
		}
		if c.Zen || c.Sound != "" || c.MoveBoxAlways || c.Pieces != "" || c.SoundAsked {
			t.Errorf("settings not reset: %+v", c)
		}
		if shownPanel() != "onboard" || Root.App.GetFocus() != onboardList {
			t.Errorf("after reset: on %s, want the first-run setup", shownPanel())
		}
		gotoWelcome()
	})
}

// First-run setup offers Stockfish when there's none, once.
func TestOnboardingOffersStockfish(t *testing.T) {
	scr := startTestApp(t, 160, 50)
	if findEngine() != "" {
		t.Skip("Stockfish is installed here: the step is skipped")
	}
	onUI(func() {
		_ = api.UpdateConfig(func(c *api.UserConfig) {
			c.Pieces, c.SoundAsked, c.KeysAsked, c.StockfishAsked = PiecesText, true, true, false
		})
		runOnboarding()
	})
	waitFor(t, "the Stockfish question", func() bool {
		return shownPanel() == "onboard" && onboardList.GetItemCount() == 3 &&
			strings.HasPrefix(onboardList.GetItem(0).GetMainText(), "Download Stockfish")
	})
	press(scr, tc.KeyRune, "3") // Not now
	waitFor(t, "home", func() bool { return shownPanel() == "welcome" })
	onUI(func() {
		if !api.UserInfo.StockfishAsked {
			t.Error("not remembered as asked")
		}
		if needsOnboarding() {
			t.Error("setup would show again")
		}
	})
}

// First-run setup asks arrows or hjkl; arrows only frees h j k l: the
// board ignores them, lists don't move on them, and the ribbons stop
// mentioning them.
func TestOnboardingKeys(t *testing.T) {
	scr := startTestApp(t, 160, 50)
	defer onUI(func() { setVimKeys(true) })
	onUI(func() {
		_ = api.UpdateConfig(func(c *api.UserConfig) {
			c.Pieces, c.SoundAsked, c.KeysAsked, c.StockfishAsked = PiecesText, true, false, true
		})
		runOnboarding()
	})
	waitFor(t, "the keys question", func() bool {
		return shownPanel() == "onboard" && strings.HasPrefix(onboardList.GetItem(1).GetMainText(), "Arrow keys only")
	})
	press(scr, tc.KeyRune, "2")
	waitFor(t, "home", func() bool { return shownPanel() == "welcome" })
	onUI(func() {
		if vimKeys() || !api.UserInfo.KeysAsked {
			t.Fatalf("hjkl still on (asked %v)", api.UserInfo.KeysAsked)
		}
		b := NewBoardView(func(chess.Square) {})
		b.InputHandler()(tc.NewEventKey(tc.KeyRune, "k", tc.ModNone), func(cv.Primitive) {})
		if b.keyboard {
			t.Error("k still moves on the board")
		}
		for tv := range keyRibbons {
			if strings.Contains(tv.GetText(false), "hjkl") {
				t.Errorf("ribbon still says hjkl: %q", tv.GetText(false))
			}
		}
		if len(cv.Keys.MoveUp2) != 0 {
			t.Error("lists still move on k")
		}
		setVimKeys(true)
		if len(cv.Keys.MoveUp2) == 0 || boardKeys() != "hjkl[] " {
			t.Error("turning hjkl back on didn't")
		}
	})
}
