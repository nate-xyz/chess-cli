package pkg

import "testing"

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
