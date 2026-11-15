//go:build windows

package pkg

import (
	"sync"
	"syscall"
	"unsafe"
)

// Playing a WAV on Windows with winmm's PlaySound, from memory: built into
// Windows, no cgo needed.

var procPlaySound = syscall.NewLazyDLL("winmm.dll").NewProc("PlaySoundW")

const (
	sndAsync     = 0x0001
	sndNoDefault = 0x0002
	sndMemory    = 0x0004
)

func soundAvailable() bool { return procPlaySound.Find() == nil }

// playing keeps the sound being played alive (PlaySound reads it while it
// plays asynchronously).
var (
	playingMu sync.Mutex
	playing   []byte
)

func playWAV(_ string, data []byte) error {
	playingMu.Lock()
	defer playingMu.Unlock()
	playing = data
	r, _, err := procPlaySound.Call(uintptr(unsafe.Pointer(&playing[0])), 0, sndAsync|sndNoDefault|sndMemory)
	if r == 0 {
		return err
	}
	return nil
}

func soundHelp() string { return "Windows sound isn't available" }
