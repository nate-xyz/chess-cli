package pkg

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tc "github.com/gdamore/tcell/v3"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
)

func TestStudyScreen(t *testing.T) {
	startTestApp(t, 160, 50)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/study/abcd1234.pgn" || r.URL.Query().Get("variations") != "true" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(testStudyPGN))
	}))
	defer srv.Close()
	oldHost := api.HostUrl
	api.HostUrl = srv.URL
	defer func() { api.HostUrl = oldHost }()

	ss := Root.study
	onUI(func() { openStudy("abcd1234", "") })
	waitFor(t, "study loaded", func() bool { return ss.ch != nil })
	onUI(func() {
		if ss.ch.name() != "Italian" || ss.white {
			t.Errorf("chapter %q, white at bottom %v (the study says black)", ss.ch.name(), ss.white)
		}
		if ss.Chapters.GetItemCount() != 3 {
			t.Errorf("%d chapters listed", ss.Chapters.GetItemCount())
		}
		if !strings.Contains(ss.Info.GetText(true), "Openings") {
			t.Errorf("info = %q", ss.Info.GetText(true))
		}
		// the start: the chapter's circle on e4
		if len(ss.Board.shapes) != 1 {
			t.Errorf("start shapes = %v", ss.Board.shapes)
		}

		// → steps along the main line
		ss.step(3) // 1. e4 e5 2. Nf3
		if ss.cur.san != "Nf3" || len(ss.Board.shapes) != 2 {
			t.Errorf("after 3 steps: %q, shapes %v", ss.cur.san, ss.Board.shapes)
		}
		if hl := ss.Moves.GetHighlights(); len(hl) != 1 || hl[0] != "n3" {
			t.Errorf("highlight = %v", hl)
		}

		// clicking a move in a variation switches to that line
		ss.Moves.Highlight("n8") // 3... Nf6 in the 2... d6 line
		if ss.cur.san != "Nf6" || ss.cur.parent.san != "d4" {
			t.Fatalf("clicked into the variation, at %q", ss.cur.san)
		}
		ss.Board.mu.Lock()
		pos, _, to := ss.Board.shown()
		ss.Board.mu.Unlock()
		if pos.String() != ss.cur.pos.String() || to != chess.F6 {
			t.Errorf("board shows %v", pos)
		}
		ss.step(-1)
		if ss.cur.san != "d4" {
			t.Errorf("← from Nf6 = %q, want d4", ss.cur.san)
		}
		ss.step(-100)
		if ss.cur != ss.ch.root {
			t.Error("↑ should go to the start")
		}

		// the players' lines stay while stepping back (no history banner in studies)
		ss.Board.mu.Lock()
		browse, back := ss.Board.browse, ss.Board.back
		ss.Board.mu.Unlock()
		if !browse || back == 0 {
			t.Errorf("browse %v back %v", browse, back)
		}

		// another chapter, from a position
		ss.selectChapter(1)
		if ss.cur != ss.ch.root || ss.ch.name() != "Endgame" || !ss.white {
			t.Errorf("chapter 2: %q", ss.ch.name())
		}
		ss.Board.mu.Lock()
		pos, _, _ = ss.Board.shown()
		ss.Board.mu.Unlock()
		if !strings.HasPrefix(pos.String(), "8/8/8/4k3/8/8/4P3/4K3") {
			t.Errorf("endgame start = %v", pos)
		}

		// a chapter with a move that can't be read says so
		ss.selectChapter(2)
		if !strings.Contains(ss.Moves.GetText(true), "Stopped reading") {
			t.Error("no warning about the unreadable move")
		}
	})

	// the wheel / [ ] on the board moves the highlight and the shapes
	onUI(func() {
		ss.selectChapter(0)
		ss.Board.ShowPly(3)
	})
	waitFor(t, "board stepping updates the study", func() bool {
		return ss.cur != nil && ss.cur.san == "Nf3" && len(ss.Moves.GetHighlights()) == 1
	})
}

func TestStudyIDFrom(t *testing.T) {
	for in, want := range map[string][2]string{
		"abcd1234":                                    {"abcd1234", ""},
		"https://lichess.org/study/abcd1234":          {"abcd1234", ""},
		"https://lichess.org/study/abcd1234/efgh5678": {"abcd1234", "efgh5678"},
		"lichess.org/study/abcd1234/efgh5678#3":       {"abcd1234", "efgh5678"},
	} {
		s, c := api.StudyIDFrom(in)
		if s != want[0] || c != want[1] {
			t.Errorf("StudyIDFrom(%q) = %q, %q", in, s, c)
		}
	}
	if api.ValidStudyID("../../x") || api.ValidStudyID("abc") {
		t.Error("bad study IDs accepted")
	}
}

// Side lines branching at the current move are underlined, heavier the
// deeper they're nested: solid, then dashed.
func TestStudyUnderlines(t *testing.T) {
	scr := startTestApp(t, 160, 50)
	ss := Root.study
	onUI(func() {
		Root.Switch("study")
		ss.show(parseStudyPGN(testStudyPGN), "")
		ss.goTo(ss.ch.nodes[4]) // 2... Nc6, where 2... d6 branches off
	})
	waitFor(t, "drawn", func() bool { Root.App.Draw(); return true })
	underlineAt := func(text string) tc.UnderlineStyle {
		w, h := scr.Size()
		for y := 0; y < h; y++ {
			var row strings.Builder
			for x := 0; x < w; x++ {
				s, _, _ := scr.Get(x, y)
				row.WriteString(s)
			}
			if i := strings.Index(row.String(), text); i >= 0 {
				x := len([]rune(row.String()[:i]))
				_, style, _ := scr.Get(x+1, y)
				return style.GetUnderlineStyle()
			}
		}
		t.Fatalf("%q not on screen", text)
		return 0
	}
	var top, nested, comment tc.UnderlineStyle
	onUI(func() {
		top, nested, comment = underlineAt("(2... d6"), underlineAt("(3. Bc4)"), underlineAt("Philidor")
	})
	if top != tc.UnderlineStyleSolid || comment != tc.UnderlineStyleSolid {
		t.Errorf("first level: line %v, its comment %v, want solid", top, comment)
	}
	if nested != tc.UnderlineStyleDashed {
		t.Errorf("nested line: %v, want dashed", nested)
	}
}
