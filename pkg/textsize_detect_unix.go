//go:build unix

package pkg

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

var cprPattern = regexp.MustCompile(`\x1b\[(\d+);(\d+)R`)

// probeTerminal asks the terminal which of kitty's protocols it supports,
// in one go before the UI starts.
//
// Text sizing, as its spec recommends: report the cursor position, draw a
// space two cells wide (w=2), report again, draw a space scaled 2x (s=2),
// report again. If the cursor moved two cells each time, both parts are
// supported. Terminals that don't support it just don't move the cursor.
//
// Graphics: a query for a 1x1 image (a=q stores nothing) is sent first;
// terminals that support the protocol answer "OK" before the cursor reports,
// others ignore it.
func probeTerminal() (textSizing, graphics bool) {
	in, out := int(os.Stdin.Fd()), os.Stdout
	if !term.IsTerminal(in) || !term.IsTerminal(int(out.Fd())) {
		return false, false
	}
	old, err := term.MakeRaw(in)
	if err != nil {
		return false, false
	}
	defer term.Restore(in, old)

	_, _ = out.WriteString("\x1b_Gi=31,s=1,v=1,a=q,t=d,f=24;AAAA\x1b\\" +
		"\r\x1b[6n\x1b]66;w=2; \a\x1b[6n\x1b]66;s=2; \a\x1b[6n")
	cols, reply := readCursorReports(in, 3, 500*time.Millisecond)
	_, _ = out.WriteString("\r\x1b[J") // erase the test characters

	textSizing = len(cols) == 3 && cols[1]-cols[0] == 2 && cols[2]-cols[1] == 2
	graphics = strings.Contains(string(reply), "_Gi=31;OK")
	return textSizing, graphics
}

func readCursorColumns(fd, n int, timeout time.Duration) []int {
	cols, _ := readCursorReports(fd, n, timeout)
	return cols
}

// readCursorReports reads cursor position reports until it has n of them or
// the timeout passes, and returns their columns and everything read.
func readCursorReports(fd, n int, timeout time.Duration) ([]int, []byte) {
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
			return cols, buf
		}
		left := time.Until(deadline)
		if left <= 0 {
			return nil, buf
		}
		fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		ready, err := unix.Poll(fds, int(left.Milliseconds())+1)
		if err == unix.EINTR {
			continue
		}
		if err != nil || ready == 0 {
			return nil, buf
		}
		k, err := unix.Read(fd, chunk)
		if err != nil || k <= 0 {
			return nil, buf
		}
		buf = append(buf, chunk[:k]...)
	}
}
