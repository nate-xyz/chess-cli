package pkg

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func TestSoundsSynthesise(t *testing.T) {
	for _, theme := range []string{SoundRetro, SoundWood} {
		for ev := soundEvent(0); ev < soundEventCount; ev++ {
			b := soundData(theme, ev)
			if string(b[0:4]) != "RIFF" || string(b[8:12]) != "WAVE" || binary.LittleEndian.Uint32(b[24:28]) != sampleRate {
				t.Fatalf("%s/%s: not a %d Hz WAV", theme, soundNames[ev], sampleRate)
			}
			n := int(binary.LittleEndian.Uint32(b[40:44])) / 2
			secs := float64(n) / sampleRate
			peak := 0.0
			for i := 0; i < n; i++ {
				peak = math.Max(peak, math.Abs(float64(int16(binary.LittleEndian.Uint16(b[44+2*i:])))))
			}
			if secs < 0.03 || secs > 1 {
				t.Errorf("%s/%s lasts %.2fs", theme, soundNames[ev], secs)
			}
			if peak < 3000 || peak >= 32767 {
				t.Errorf("%s/%s peak %v: too quiet or clipping", theme, soundNames[ev], peak)
			}
		}
	}
}

func TestSoundFileCached(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	data := soundData(SoundRetro, soundMove)
	p, err := soundFile(SoundRetro, soundMove, data)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(p); len(got) != len(data) || filepath.Ext(p) != ".wav" {
		t.Errorf("cached file %s: %d bytes", p, len(got))
	}
}

func TestWhichSound(t *testing.T) {
	g := chess.NewGame()
	for _, m := range []string{"e4", "d5", "exd5", "Qxd5", "Nc3", "Qe5+"} {
		if err := g.MoveStr(m); err != nil {
			t.Fatal(err)
		}
	}
	moves := g.Moves()
	if moveSound(moves[0]) != soundMove || moveSound(moves[2]) != soundCapture || moveSound(moves[5]) != soundCheck {
		t.Error("move / capture / check sounds")
	}
	if resultSound(chess.White, chess.White) != soundVictory || resultSound(chess.White, chess.Black) != soundDefeat ||
		resultSound(chess.NoColor, chess.White) != soundDraw || resultSound(chess.Black, chess.NoColor) != soundVictory {
		t.Error("result sounds")
	}
	if !lowTime(8*time.Second, 5*time.Minute) || lowTime(20*time.Second, 5*time.Minute) || lowTime(5*time.Second, 0) {
		t.Error("low time: under 10 s, not without a clock")
	}
	if !lowTime(9*time.Second, time.Minute) || lowTime(12*time.Second, time.Minute) {
		t.Error("low time in a 1-minute game")
	}
}

func TestSoundSetting(t *testing.T) {
	startTestApp(t, 160, 50)
	onUI(func() {
		defer setSoundTheme("")
		if soundTheme() != SoundRetro {
			t.Errorf("default theme %q, want retro", soundTheme())
		}
		setSoundTheme(SoundOff)
		if soundTheme() != SoundOff || api.UserInfo.Sound != SoundOff {
			t.Error("turning sound off")
		}
	})
}

// The first-run screen asks about sound after the rendering choice.
func TestOnboardSound(t *testing.T) {
	startTestApp(t, 160, 50)
	onUI(func() {
		defer func() { api.UserInfo.Sound, api.UserInfo.SoundAsked = "", false }()
		api.UserInfo.SoundAsked = false
		info := textBox(0)
		list := menuList()
		onboardSound(list, info)
		if list.GetItemCount() != 3 {
			t.Fatalf("%d sound choices", list.GetItemCount())
		}
		list.SetCurrentItem(2) // No sound
		list.InputHandler()(tc.NewEventKey(tc.KeyEnter, "", tc.ModNone), func(cv.Primitive) {})
		if !api.UserInfo.SoundAsked || soundTheme() != SoundOff {
			t.Errorf("after choosing no sound: asked %v, theme %q", api.UserInfo.SoundAsked, soundTheme())
		}
	})
}
