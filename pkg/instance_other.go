//go:build !unix

package pkg

import "errors"

// ErrAlreadyRunning means another chess-cli is running for this user.
var ErrAlreadyRunning = errors.New("chess-cli is already running")

// LockInstance: not enforced on this platform yet.
func LockInstance(dir string) error { return nil }
