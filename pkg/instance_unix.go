//go:build unix

package pkg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// ErrAlreadyRunning means another chess-cli is running for this user.
var ErrAlreadyRunning = errors.New("chess-cli is already running")

var instanceLock *os.File // held open (and locked) for the life of the process

// LockInstance makes sure only one chess-cli runs at a time: two would both
// follow the Lichess event stream and write the same config and saved games.
// It takes an exclusive lock on a file in dir; the system releases it when
// the process exits, even after a crash, so it can't go stale.
func LockInstance(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, "chess-cli.lock"), os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		defer f.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			b := make([]byte, 32)
			n, _ := f.ReadAt(b, 0)
			if pid, err := strconv.Atoi(strings.TrimSpace(string(b[:n]))); err == nil && pid > 0 {
				return fmt.Errorf("%w (process %d, in another terminal)", ErrAlreadyRunning, pid)
			}
			return ErrAlreadyRunning
		}
		return err
	}
	_ = f.Truncate(0)
	_, _ = f.WriteAt([]byte(strconv.Itoa(os.Getpid())+"\n"), 0)
	instanceLock = f
	return nil
}
