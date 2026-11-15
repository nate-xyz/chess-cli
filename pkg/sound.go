package pkg

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

// Sound: chess-cli's own sounds, synthesised here (no audio files), in two
// themes: Retro (chiptune square waves, the default) and Wood (a softer
// piece-on-board thock). They're written once as small WAV files in the
// cache folder and played with the system's player (see sound_*.go), so
// there's no audio library and the builds stay static.

type soundEvent int

const (
	soundMove soundEvent = iota
	soundCapture
	soundCheck
	soundCastle
	soundStart   // a game starts
	soundVictory // you won (or a two-player game was decided)
	soundDefeat
	soundDraw
	soundLowTime
	soundError   // an illegal or wrong move
	soundConfirm // a right move in a puzzle
	soundEventCount
)

var soundNames = [...]string{"move", "capture", "check", "castle", "start", "victory", "defeat", "draw", "lowtime", "error", "confirm"}

// Sound themes, as saved in the config ("" is the default, Retro).
const (
	SoundRetro = "retro"
	SoundWood  = "wood"
	SoundOff   = "off"
)

var soundThemes = []string{SoundRetro, SoundWood, SoundOff}

var soundThemeNames = map[string]string{SoundRetro: "Retro", SoundWood: "Wood", SoundOff: "Off"}

// soundTheme is the chosen theme.
func soundTheme() string {
	switch t := api.UserInfo.Sound; t {
	case SoundWood, SoundOff:
		return t
	}
	return SoundRetro
}

func setSoundTheme(t string) {
	_ = api.UpdateConfig(func(c *api.UserConfig) { c.Sound = t })
}

// soundsMuted turns sound off without changing the setting (tests).
var soundsMuted atomic.Bool

var lastSound atomic.Int64 // unix nanos of the last sound started

// playSound plays a sound in the chosen theme, without waiting. Sounds
// closer together than 35 ms are dropped (e.g. holding → in a study).
func playSound(ev soundEvent) { playSoundIn(soundTheme(), ev) }

func playSoundIn(theme string, ev soundEvent) {
	if theme == SoundOff || soundsMuted.Load() {
		return
	}
	now := time.Now().UnixNano()
	if last := lastSound.Load(); now-last < int64(35*time.Millisecond) || !lastSound.CompareAndSwap(last, now) {
		return
	}
	go func() {
		data := soundData(theme, ev)
		path, err := soundFile(theme, ev, data)
		if err != nil {
			return
		}
		_ = playWAV(path, data)
	}()
}

// moveSound is the sound for a move: capture, check, castling or a move.
func moveSound(m *chess.Move) soundEvent {
	switch {
	case m == nil:
		return soundMove
	case m.HasTag(chess.Check):
		return soundCheck
	case m.HasTag(chess.Capture) || m.HasTag(chess.EnPassant):
		return soundCapture
	case m.HasTag(chess.KingSideCastle) || m.HasTag(chess.QueenSideCastle):
		return soundCastle
	}
	return soundMove
}

// resultSound is the sound at the end of a game: winner is chess.NoColor
// for a draw; me is the player's colour (chess.NoColor when two people
// play on one keyboard: any decisive result is a victory).
func resultSound(winner, me chess.Color) soundEvent {
	switch {
	case winner == chess.NoColor:
		return soundDraw
	case me == chess.NoColor || winner == me:
		return soundVictory
	}
	return soundDefeat
}

// lowTime reports whether a clock is low enough for the warning sound:
// under 10 seconds, or a tenth of the starting time in very fast games.
func lowTime(left, initial time.Duration) bool {
	return initial >= 30*time.Second && left > 0 && left < min(10*time.Second, initial/10+5*time.Second)
}

var (
	soundCache   sync.Map // theme/event -> []byte
	soundFilesMu sync.Mutex
)

func soundData(theme string, ev soundEvent) []byte {
	key := theme + "/" + soundNames[ev]
	if b, ok := soundCache.Load(key); ok {
		return b.([]byte)
	}
	b := wavBytes(synthesise(theme, ev))
	soundCache.Store(key, b)
	return b
}

// soundFile writes a sound to the cache folder (once) and returns its path.
func soundFile(theme string, ev soundEvent, data []byte) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "chess-cli", "sounds")
	// the version changes when the sounds do, so old files aren't reused
	p := filepath.Join(dir, theme+"-"+soundNames[ev]+"-v1.wav")
	soundFilesMu.Lock()
	defer soundFilesMu.Unlock()
	if st, err := os.Stat(p); err == nil && st.Size() == int64(len(data)) {
		return p, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return "", err
	}
	return p, os.Rename(tmp, p)
}

// Synthesis.

const sampleRate = 22050

type wave int

const (
	square wave = iota
	triangle
	sine
	noise
)

// note is one sound in a mix: a wave from f0 gliding to f1 Hz, starting at
// start seconds, lasting dur, with an attack and an exponential decay.
type note struct {
	start, dur float64
	f0, f1     float64
	w          wave
	amp        float64
	decay      float64 // time constant in seconds (0 = flat)
	lowpass    float64 // 0-1 smoothing for noise (0 = none)
}

func synthesise(theme string, ev soundEvent) []float64 {
	var notes []note
	if theme == SoundWood {
		notes = woodSound(ev)
	} else {
		notes = retroSound(ev)
	}
	end := 0.0
	for _, n := range notes {
		end = math.Max(end, n.start+n.dur)
	}
	out := make([]float64, int((end+0.01)*sampleRate))
	rng := rand.New(rand.NewPCG(uint64(ev)+1, 7)) // the same noise every time
	for _, n := range notes {
		i0 := int(n.start * sampleRate)
		count := int(n.dur * sampleRate)
		phase, lp := 0.0, 0.0
		for i := 0; i < count && i0+i < len(out); i++ {
			t := float64(i) / sampleRate
			f := n.f0 + (n.f1-n.f0)*t/n.dur
			phase += f / sampleRate
			var v float64
			switch n.w {
			case square:
				v = 1
				if math.Mod(phase, 1) >= 0.5 {
					v = -1
				}
			case triangle:
				v = 4*math.Abs(math.Mod(phase, 1)-0.5) - 1
			case sine:
				v = math.Sin(2 * math.Pi * phase)
			case noise:
				v = rng.Float64()*2 - 1
				if n.lowpass > 0 {
					lp += (v - lp) * (1 - n.lowpass)
					v = lp * 2
				}
			}
			env := math.Min(1, t/0.002) // 2 ms attack: no click
			if n.decay > 0 {
				env *= math.Exp(-t / n.decay)
			}
			env *= math.Min(1, (n.dur-t)/0.004) // and a 4 ms release
			out[i0+i] += v * n.amp * env
		}
	}
	return out
}

// Notes in Hz.
const (
	c5, d5, e5, f5, g5, a5, b5 = 523.25, 587.33, 659.25, 698.46, 783.99, 880.0, 987.77
	c6, e6, g6                 = 1046.5, 1318.5, 1568.0
	bb4, eb5                   = 466.16, 622.25
)

// retroSound: square-wave blips, like an 8-bit console.
func retroSound(ev soundEvent) []note {
	blip := func(start, f0, f1, dur, amp float64) note {
		return note{start: start, dur: dur, f0: f0, f1: f1, w: square, amp: amp, decay: dur / 2}
	}
	arp := func(freqs []float64, step, last float64) []note {
		var ns []note
		for i, f := range freqs {
			d := step
			if i == len(freqs)-1 {
				d = last
			}
			ns = append(ns, note{start: float64(i) * step, dur: d, f0: f, f1: f, w: square, amp: 0.16, decay: d * 0.7})
		}
		return ns
	}
	switch ev {
	case soundMove:
		return []note{blip(0, 700, 520, 0.05, 0.2)}
	case soundCapture:
		return []note{
			{start: 0, dur: 0.06, w: noise, amp: 0.28, decay: 0.025},
			blip(0, 420, 200, 0.08, 0.2),
		}
	case soundCheck:
		return []note{blip(0, 990, 990, 0.05, 0.18), blip(0.06, 1320, 1320, 0.07, 0.18)}
	case soundCastle:
		return []note{blip(0, 700, 520, 0.05, 0.2), blip(0.07, 620, 460, 0.05, 0.2)}
	case soundStart:
		return arp([]float64{c5, e5, g5, c6}, 0.07, 0.16)
	case soundVictory:
		return arp([]float64{g5, c6, e6, g6}, 0.08, 0.3)
	case soundDefeat:
		return arp([]float64{c5, bb4, 392, 311.13}, 0.11, 0.3)
	case soundDraw:
		return arp([]float64{e5, e5}, 0.13, 0.22)
	case soundLowTime:
		return []note{blip(0, 1760, 1760, 0.06, 0.14), blip(0.12, 1760, 1760, 0.06, 0.14)}
	case soundError:
		return []note{{start: 0, dur: 0.14, f0: 140, f1: 110, w: square, amp: 0.16, decay: 0.1}}
	case soundConfirm:
		return arp([]float64{g5, c6}, 0.06, 0.12)
	}
	return nil
}

// woodSound: a soft thock for moves, gentle chimes for the rest.
func woodSound(ev soundEvent) []note {
	thock := func(start, amp float64) []note {
		return []note{
			{start: start, dur: 0.09, f0: 210, f1: 150, w: sine, amp: amp, decay: 0.025},
			{start: start, dur: 0.02, w: noise, amp: amp * 0.7, decay: 0.006, lowpass: 0.6},
		}
	}
	chime := func(freqs []float64, step float64) []note {
		var ns []note
		for i, f := range freqs {
			ns = append(ns,
				note{start: float64(i) * step, dur: 0.45, f0: f, f1: f, w: sine, amp: 0.16, decay: 0.15},
				note{start: float64(i) * step, dur: 0.3, f0: 2 * f, f1: 2 * f, w: sine, amp: 0.04, decay: 0.08})
		}
		return ns
	}
	switch ev {
	case soundMove:
		return thock(0, 0.5)
	case soundCapture:
		return append(thock(0, 0.55), thock(0.045, 0.45)...)
	case soundCheck:
		return append(thock(0, 0.5), chime([]float64{e6}, 0)...)
	case soundCastle:
		return append(thock(0, 0.45), thock(0.09, 0.45)...)
	case soundStart:
		return chime([]float64{e5, b5}, 0.12)
	case soundVictory:
		return chime([]float64{c5, e5, g5, c6}, 0.11)
	case soundDefeat:
		return chime([]float64{g5, e5, c5}, 0.14)
	case soundDraw:
		return chime([]float64{d5, d5}, 0.16)
	case soundLowTime:
		return chime([]float64{a5, a5}, 0.14)
	case soundError:
		return []note{
			{start: 0, dur: 0.16, f0: 120, f1: 100, w: sine, amp: 0.35, decay: 0.06},
			{start: 0, dur: 0.05, w: noise, amp: 0.15, decay: 0.02, lowpass: 0.8},
		}
	case soundConfirm:
		return chime([]float64{g5, d5 * 2}, 0.08)
	}
	return nil
}

// wavBytes encodes samples (-1..1) as a 16-bit mono WAV file.
func wavBytes(samples []float64) []byte {
	var b bytes.Buffer
	data := len(samples) * 2
	w := func(v any) { _ = binary.Write(&b, binary.LittleEndian, v) }
	b.WriteString("RIFF")
	w(uint32(36 + data))
	b.WriteString("WAVEfmt ")
	w(uint32(16))
	w(uint16(1)) // PCM
	w(uint16(1)) // mono
	w(uint32(sampleRate))
	w(uint32(sampleRate * 2))
	w(uint16(2))
	w(uint16(16))
	b.WriteString("data")
	w(uint32(data))
	for _, s := range samples {
		w(int16(math.Max(-1, math.Min(1, s)) * 32000))
	}
	return b.Bytes()
}
