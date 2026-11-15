package pkg

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	tc "github.com/gdamore/tcell/v3"
	"github.com/notnil/chess"
)

// captureScreen is a real (mock-terminal) screen whose raw tty writes are
// also recorded, so tests can see the escape codes sent for big pieces.
type captureScreen struct {
	tc.Screen
	tty *captureTty
}

type captureTty struct {
	tc.Tty
	mu  sync.Mutex
	buf bytes.Buffer
}

func (c *captureTty) Write(p []byte) (int, error) {
	c.mu.Lock()
	c.buf.Write(p)
	c.mu.Unlock()
	return c.Tty.Write(p)
}

func (c *captureTty) take() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.buf.String()
	c.buf.Reset()
	return s
}

func (c captureScreen) Tty() (tc.Tty, bool) { return c.tty, true }

func newCaptureScreen(t *testing.T, w, h int) captureScreen {
	s := newTestScreen(t, w, h)
	real, _ := s.Tty()
	return captureScreen{Screen: s, tty: &captureTty{Tty: real}}
}

func withTextSizing(t *testing.T) {
	textSizing.Lock()
	old := textSizing.enabled
	textSizing.enabled = true
	textSizing.screens = nil // each test's screen starts with nothing on it
	textSizing.Unlock()
	t.Cleanup(func() {
		textSizing.Lock()
		textSizing.enabled = old
		textSizing.screens = nil
		textSizing.Unlock()
	})
}

func TestSizedGeometry(t *testing.T) {
	for _, sz := range [][2]int{{40, 12}, {80, 26}, {160, 50}, {300, 120}} {
		sqW, sqH := boardGeometry(sz[0], sz[1], styleSized, 2)
		if sqH < 1 || sqH > 7 || sqW != 2*sqH {
			t.Errorf("%v: square %dx%d, want 2s x s with 1 <= s <= 7", sz, sqW, sqH)
		}
		if 8*sqW+labelW > sz[0] || 8*sqH+labelH > sz[1] {
			t.Errorf("%v: board overflows", sz)
		}
	}
	if w, _ := boardGeometry(300, 120, styleSized, 2); w != 14 {
		t.Errorf("scale should cap at 7 (square 14 wide), got %d", w)
	}
}

func TestEncodeSized(t *testing.T) {
	got := encodeSized([]sizedCell{
		{x: 4, y: 2, s: 3, w: 2, n: 9, d: 10, fg: colWhiteBody, bg: colLight, text: "♚"},
		{x: 10, y: 2, s: 3, w: 2, fg: colWhiteBody, bg: colDark, text: " "},
	})
	want := "\x1b7" +
		"\x1b[3;5H\x1b[0;38;2;255;248;220;48;2;145;130;109m\x1b]66;s=3:w=2:n=9:d=10:v=2:h=2;♚\a" +
		"\x1b[3;11H\x1b[0;38;2;255;248;220;48;2;108;81;59m\x1b]66;s=3:w=2; \a" +
		"\x1b8"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestTextSizingFlush(t *testing.T) {
	withTextSizing(t)
	s := newCaptureScreen(t, 120, 40)
	b := NewBoardView(nil)
	b.SetRect(0, 0, 120, 40)
	draw := func() string {
		b.Draw(s)
		flushTextSizing(s)
		return s.tty.take()
	}

	out := draw()
	if n := strings.Count(out, "\x1b]66;"); n != 64 {
		t.Fatalf("first frame sent %d multicells, want 64", n)
	}
	// pieces carry a trailing space so kitty centres them (see squareLook.sized)
	if !strings.Contains(out, ";♚ \a") || !strings.Contains(out, ";♟\uFE0E \a") {
		t.Errorf("pieces missing or unpadded in %q", out[:200])
	}
	if out := draw(); out != "" {
		t.Errorf("unchanged board was re-sent: %q", out)
	}

	g := chess.NewGame()
	b.HandleSquare(g, chess.E2, true, func(string) {})
	if out := draw(); strings.Count(out, "\x1b]66;") != 64 {
		t.Errorf("selection change wasn't sent")
	}

	// board hidden (e.g. switched screens): multicells erased, nothing new sent
	flushTextSizing(s)
	out = s.tty.take()
	if strings.Contains(out, "\x1b]66;") || strings.Count(out, "\x1b[0m ") != 64 {
		t.Errorf("hiding the board should erase its 64 multicells, got %q", out)
	}
}

// Two screens keep separate multicell state: drawing a board on one must not
// make the other erase or resend anything (in tests the background app
// redraws its own screen while a test draws on another).
func TestTextSizingPerScreen(t *testing.T) {
	withTextSizing(t)
	a, other := newCaptureScreen(t, 120, 40), newCaptureScreen(t, 160, 50)
	ba, bo := NewBoardView(nil), NewBoardView(nil)
	ba.SetRect(0, 0, 120, 40)
	bo.SetRect(0, 0, 160, 50)
	ba.Draw(a)
	flushTextSizing(a)
	a.tty.take()
	bo.Draw(other) // another screen draws in between
	flushTextSizing(other)
	ba.Draw(a)
	flushTextSizing(a)
	if out := a.tty.take(); out != "" {
		t.Errorf("the other screen's board leaked into this one: %q", out[:min(len(out), 120)])
	}
}
