//go:build unix

package pkg

import (
	"os"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

var cprPattern = regexp.MustCompile(`\x1b\[(\d+);(\d+)R`)

// detectTextSizing asks the terminal whether it supports kitty's text sizing
// protocol, as the spec recommends: report the cursor position, draw a space
// two cells wide (w=2), report again, draw a space scaled 2x (s=2), report
// again. If the cursor moved two cells each time, both parts are supported.
// Terminals that don't support it just don't move the cursor.
func detectTextSizing() bool {
	in, out := int(os.Stdin.Fd()), os.Stdout
	if !term.IsTerminal(in) || !term.IsTerminal(int(out.Fd())) {
		return false
	}
	old, err := term.MakeRaw(in)
	if err != nil {
		return false
	}
	defer term.Restore(in, old)

	_, _ = out.WriteString("\r\x1b[6n\x1b]66;w=2; \a\x1b[6n\x1b]66;s=2; \a\x1b[6n")
	cols := readCursorColumns(in, 3, 500*time.Millisecond)
	_, _ = out.WriteString("\r\x1b[J") // erase the test characters

	return len(cols) == 3 && cols[1]-cols[0] == 2 && cols[2]-cols[1] == 2
}

// readCursorColumns reads cursor position reports until it has n of them or
// the timeout passes, and returns their columns.
func readCursorColumns(fd, n int, timeout time.Duration) []int {
	deadline := time.Now().Add(timeout)
	var buf []byte
	chunk := make([]byte, 256)
	for {
		matches := cprPattern.FindAllSubmatch(buf, -1)
		if len(matches) >= n {
			cols := make([]int, 0, n)
			for _, m := range matches[:n] {
				c, _ := strconv.Atoi(string(m[2]))
				cols = append(cols, c)
			}
			return cols
		}
		left := time.Until(deadline)
		if left <= 0 {
			return nil
		}
		fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		ready, err := unix.Poll(fds, int(left.Milliseconds())+1)
		if err == unix.EINTR {
			continue
		}
		if err != nil || ready == 0 {
			return nil
		}
		k, err := unix.Read(fd, chunk)
		if err != nil || k <= 0 {
			return nil
		}
		buf = append(buf, chunk[:k]...)
	}
}
