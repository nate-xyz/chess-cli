//go:build unix

package pkg

import (
	"os"
	"testing"
	"time"
)

func TestReadCursorColumns(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	// kitty: the cursor moves 2 cells after each test character
	go w.WriteString("\x1b[12;1R\x1b[12;3R\x1b[1")
	go func() { time.Sleep(20 * time.Millisecond); w.WriteString("2;5R") }()
	cols := readCursorColumns(int(r.Fd()), 3, time.Second)
	if len(cols) != 3 || cols[0] != 1 || cols[1] != 3 || cols[2] != 5 {
		t.Fatalf("cols = %v", cols)
	}

	// a terminal that never answers: give up at the timeout
	r2, w2, _ := os.Pipe()
	defer r2.Close()
	defer w2.Close()
	start := time.Now()
	if cols := readCursorColumns(int(r2.Fd()), 3, 100*time.Millisecond); cols != nil {
		t.Fatalf("expected nil, got %v", cols)
	}
	if time.Since(start) > time.Second {
		t.Fatal("timeout not respected")
	}
}
