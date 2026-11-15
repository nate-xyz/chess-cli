//go:build unix

package pkg

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

// The lock is per process, so the second instance is a child process.
func TestLockInstance(t *testing.T) {
	if dir := os.Getenv("LOCK_TEST_DIR"); dir != "" { // the child
		if err := LockInstance(dir); errors.Is(err, ErrAlreadyRunning) {
			os.Exit(3)
		}
		os.Exit(0)
	}
	dir := t.TempDir()
	if err := LockInstance(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { instanceLock.Close(); instanceLock = nil }()
	cmd := exec.Command(os.Args[0], "-test.run=TestLockInstance")
	cmd.Env = append(os.Environ(), "LOCK_TEST_DIR="+dir)
	err := cmd.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 {
		t.Fatalf("second instance wasn't refused: %v", err)
	}
	instanceLock.Close() // first instance exits: the lock is free again
	instanceLock = nil
	cmd = exec.Command(os.Args[0], "-test.run=TestLockInstance")
	cmd.Env = append(os.Environ(), "LOCK_TEST_DIR="+dir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("lock not released: %v", err)
	}
}
