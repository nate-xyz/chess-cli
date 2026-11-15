package pkg

import (
	"strings"
	"testing"

	tc "github.com/gdamore/tcell/v3"
)

// Studies in each rendering mode: full terminals get the greys, tints and
// styled underlines; basic rendering (the original chess-cli look, for old
// terminals) only standard colours and plain underlines, and the move
// symbols still show on the board.
func TestStudyRenderingModes(t *testing.T) {
	scr := startTestApp(t, 160, 50)
	ss := Root.study
	defer onUI(func() { previewPieces(""); applyUIColours() })

	find := func(text string) (x, y int, ok bool) {
		w, h := scr.Size()
		for y := 0; y < h; y++ {
			var row strings.Builder
			for x := 0; x < w; x++ {
				s, _, _ := scr.Get(x, y)
				row.WriteString(s)
			}
			if i := strings.Index(row.String(), text); i >= 0 {
				return len([]rune(row.String()[:i])), y, true
			}
		}
		return 0, 0, false
	}
	styleAt := func(text string, dx int) tc.Style {
		x, y, ok := find(text)
		if !ok {
			t.Fatalf("%q not on screen", text)
		}
		_, st, _ := scr.Get(x+dx, y)
		return st
	}
	draw := func() {
		onUI(func() { Root.App.Draw() })
		onUI(func() {}) // let queued colour changes run
		onUI(func() { Root.App.Draw() })
	}

	for _, mode := range []string{PiecesText, PiecesBasic} {
		t.Run(mode, func(t *testing.T) {
			basic := mode == PiecesBasic
			onUI(func() {
				previewPieces(mode)
				Root.Switch("study")
				ss.show(parseStudyPGN(testStudyPGN), "")
			})
			draw()
			onUI(func() { ss.goTo(ss.ch.nodes[4]) }) // 2... Nc6: side lines branch here
			draw()
			var cur, nested tc.Style
			onUI(func() { cur, nested = styleAt("Nc6", 0), styleAt("(3. Bc4)", 1) })
			wantBg := colCurrentMove
			if basic {
				wantBg = tc.ColorSilver
			}
			if cur.GetBackground() != wantBg {
				t.Errorf("current move background %v, want %v", cur.GetBackground(), wantBg)
			}
			wantUl := tc.UnderlineStyleDashed
			if basic {
				wantUl = tc.UnderlineStyleSolid
			}
			if nested.GetUnderlineStyle() != wantUl || nested.HasBlink() || nested.HasStrikeThrough() {
				t.Errorf("nested side line underline %v (blink %v), want %v", nested.GetUnderlineStyle(), nested.HasBlink(), wantUl)
			}

			onUI(func() { ss.goTo(ss.ch.nodes[3]) }) // 2. Nf3: has a comment
			draw()
			var comment tc.Style
			onUI(func() { comment = styleAt("Attacks e5.", 0) })
			tinted := comment.GetBackground() == tc.NewHexColor(0x1e2a1e)
			if tinted == basic {
				t.Errorf("comment tinted %v in %s mode", tinted, mode)
			}

			// the move symbol on the board: 3. Bc4!? on c4
			onUI(func() { ss.goTo(ss.ch.nodes[9]) })
			draw()
			onUI(func() {
				b := ss.Board
				b.mu.Lock()
				row, col := displayOf(b.glyphSq, b.white)
				x, y := b.ox+(col+1)*b.sqW-2, b.oy+row*b.sqH
				b.mu.Unlock()
				s1, st, _ := scr.Get(x, y)
				s2, _, _ := scr.Get(x+1, y)
				want := moveGlyph(5).tc()
				if basic {
					want = moveGlyph(5).basic()
				}
				if s1+s2 != "!?" || st.GetBackground() != want {
					t.Errorf("badge %q on %v, want `!?` on %v", s1+s2, st.GetBackground(), want)
				}
			})
		})
	}
}
