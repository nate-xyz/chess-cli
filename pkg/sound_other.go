//go:build !windows

package pkg

import (
	"errors"
	"os/exec"
	"runtime"
	"sync"
)

// Playing a WAV file with the system's player: afplay on macOS; on Linux
// and the BSDs PipeWire's pw-play, PulseAudio's paplay or ALSA's aplay,
// whichever is installed.

var soundPlayer = sync.OnceValue(func() []string {
	candidates := [][]string{{"pw-play"}, {"paplay"}, {"aplay", "-q"}}
	if runtime.GOOS == "darwin" {
		candidates = [][]string{{"afplay"}}
	}
	for _, c := range candidates {
		if p, err := exec.LookPath(c[0]); err == nil {
			return append([]string{p}, c[1:]...)
		}
	}
	return nil
})

// soundAvailable reports whether chess-cli can play sound here.
func soundAvailable() bool { return soundPlayer() != nil }

func playWAV(path string, _ []byte) error {
	player := soundPlayer()
	if player == nil {
		return errors.New("no audio player found")
	}
	cmd := exec.Command(player[0], append(player[1:], path)...)
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}

// soundHelp says what to install when there's no player.
func soundHelp() string {
	if runtime.GOOS == "darwin" {
		return "afplay not found"
	}
	return "no audio player found (install pipewire, pulseaudio-utils or alsa-utils)"
}
