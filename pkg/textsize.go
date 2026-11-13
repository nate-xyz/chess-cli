package pkg

import (
	"fmt"
	"os"
	"strings"
	"sync"

	tc "github.com/gdamore/tcell/v3"
)

// Big pieces via kitty's text sizing protocol
// (https://sw.kovidgoyal.net/kitty/text-sizing-protocol/): the terminal draws
// a character scaled up to fill an s*w by s block of cells, in its own font,
// so pieces are real Unicode glyphs, just bigger.
//
// tcell doesn't know about these "multicell" characters, so the board locks
// its cells in tcell (tcell won't draw there; the same mechanism tcell
// provides for sixel images) and the pieces are written straight to the
// terminal after each frame, but only when something on the board changed.

// sizedCell is one multicell character: text drawn in a (s*w) x s block at x,y.
type sizedCell struct {
	x, y, s, w int
	n, d       int // fractional scale within the block (0 = none)
	fg, bg     tc.Color
	text       string
}

type rect struct{ x, y, w, h int }

var textSizing struct {
	sync.Mutex
	enabled bool

	frame   []sizedCell // collected by board Draw calls this frame
	regions []rect

	prevCells   []sizedCell // what's on the terminal now
	prevRegions []rect
	prevW       int
	prevH       int
}

// TextSizingMode reports whether big pieces are on.
func TextSizingMode() bool {
	textSizing.Lock()
	defer textSizing.Unlock()
	return textSizing.enabled
}

// SetupTextSizing decides whether to use big pieces. Call it before the UI
// takes over the terminal. CHESS_CLI_BIG_PIECES=0 turns them off, =1 forces
// them on; otherwise the terminal is asked whether it supports the protocol.
func SetupTextSizing() {
	on := false
	switch strings.ToLower(os.Getenv("CHESS_CLI_BIG_PIECES")) {
	case "0", "off", "false", "no":
	case "1", "on", "true", "yes":
		on = true
	default:
		on = detectTextSizing()
	}
	textSizing.Lock()
	textSizing.enabled = on
	textSizing.Unlock()
}

// queueSized records a board region and its multicell characters for this
// frame and locks the region so tcell leaves it alone. Called from Draw.
func queueSized(screen tc.Screen, r rect, cells []sizedCell) {
	screen.LockRegion(r.x, r.y, r.w, r.h, true)
	textSizing.Lock()
	textSizing.regions = append(textSizing.regions, r)
	textSizing.frame = append(textSizing.frame, cells...)
	textSizing.Unlock()
}

// flushTextSizing runs after every full frame (cview's after-draw hook). It
// erases multicells from boards that are gone or moved, lets tcell repaint
// those cells, and (re)sends the board's multicells if anything changed.
func flushTextSizing(screen tc.Screen) {
	ts := &textSizing
	ts.Lock()
	defer ts.Unlock()
	if !ts.enabled {
		return
	}
	frame, regions := ts.frame, ts.regions
	ts.frame, ts.regions = nil, nil
	tty, ok := screen.Tty()
	if !ok {
		return
	}

	var out strings.Builder
	out.WriteString("\x1b7") // save cursor position and attributes (tcell tracks both)

	// Boards no longer drawn here: erase their multicells (writing a space on
	// a multicell's top-left cell erases the whole block), then unlock so
	// tcell repaints whatever is there now.
	removed := false
	for _, pr := range ts.prevRegions {
		if containsRect(regions, pr) {
			continue
		}
		removed = true
		for _, c := range ts.prevCells {
			if c.x >= pr.x && c.y >= pr.y && c.x < pr.x+pr.w && c.y < pr.y+pr.h {
				fmt.Fprintf(&out, "\x1b[%d;%dH\x1b[0m ", c.y+1, c.x+1)
			}
		}
		screen.LockRegion(pr.x, pr.y, pr.w, pr.h, false)
	}
	for _, r := range regions { // an unlock above may have overlapped a current board
		screen.LockRegion(r.x, r.y, r.w, r.h, true)
	}
	out.WriteString("\x1b8")
	if removed {
		_, _ = tty.Write([]byte(out.String()))
	}
	screen.Show() // flush tcell's own drawing first; the multicells go on top

	w, h := screen.Size()
	if removed || w != ts.prevW || h != ts.prevH || !sameCells(frame, ts.prevCells) {
		_, _ = tty.Write([]byte(encodeSized(frame)))
	}
	ts.prevCells, ts.prevRegions, ts.prevW, ts.prevH = frame, regions, w, h
}

// encodeSized renders multicell characters as escape codes, leaving the
// cursor and attributes as they were.
func encodeSized(cells []sizedCell) string {
	if len(cells) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\x1b7")
	for _, c := range cells {
		fr, fg, fb := c.fg.RGB()
		br, bgc, bb := c.bg.RGB()
		meta := fmt.Sprintf("s=%d:w=%d", c.s, c.w)
		if c.d > 0 {
			meta += fmt.Sprintf(":n=%d:d=%d:v=2:h=2", c.n, c.d)
		}
		fmt.Fprintf(&b, "\x1b[%d;%dH\x1b[0;38;2;%d;%d;%d;48;2;%d;%d;%dm\x1b]66;%s;%s\a",
			c.y+1, c.x+1, fr, fg, fb, br, bgc, bb, meta, c.text)
	}
	b.WriteString("\x1b8")
	return b.String()
}

func sameCells(a, b []sizedCell) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func containsRect(rs []rect, r rect) bool {
	for _, x := range rs {
		if x == r {
			return true
		}
	}
	return false
}
