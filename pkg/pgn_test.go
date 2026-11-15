package pkg

import (
	"regexp"
	"strings"
	"testing"

	"github.com/notnil/chess"
)

// A Lichess study export: two chapters, variations, comments, symbols,
// drawings, a chapter from a position, and a chapter with a bad move.
const testStudyPGN = `[Event "Openings: Italian"]
[Site "https://lichess.org/study/abcd1234/ch000001"]
[Result "*"]
[StudyName "Openings"]
[ChapterName "Italian"]
[ChapterURL "https://lichess.org/study/abcd1234/ch000001"]
[Orientation "black"]

{ The Italian Game. [%csl Ge4] }
1. e4 e5 2. Nf3 { Attacks e5. [%cal Gf3e5,Rb8c6] } 2... Nc6 (2... d6 { Philidor } 3. d4 (3. Bc4) 3... Nf6) 3. Bc4!? $14 Bc5 (3... Nf6?! 4. Ng5) *

[Event "Openings: Endgame"]
[ChapterName "Endgame"]
[FEN "8/8/8/4k3/8/8/4P3/4K3 w - - 0 1"]
[SetUp "1"]

1. Kd2 Kd4 2. e3+ Ke4 *

[Event "Openings: Broken"]
[ChapterName "Broken"]

1. e4 e5 2. Ke3 Nc6 3. d4 *
`

func TestParseStudyPGN(t *testing.T) {
	chs := parseStudyPGN(testStudyPGN)
	if len(chs) != 3 {
		t.Fatalf("got %d chapters, want 3", len(chs))
	}
	c := chs[0]
	if c.name() != "Italian" || c.chapterID() != "ch000001" || c.white {
		t.Errorf("chapter: name %q id %q white %v", c.name(), c.chapterID(), c.white)
	}
	if c.problem != "" {
		t.Errorf("unexpected problem: %s", c.problem)
	}
	if c.root.comment != "The Italian Game." || len(c.root.shapes) != 1 || c.root.shapes[0] != (boardShape{chess.E4, chess.E4, 0}) {
		t.Errorf("root comment %q shapes %v", c.root.comment, c.root.shapes)
	}
	// main line
	var main []string
	for _, n := range c.root.line()[1:] {
		main = append(main, n.san)
	}
	if got := strings.Join(main, " "); got != "e4 e5 Nf3 Nc6 Bc4 Bc5" {
		t.Errorf("main line = %q", got)
	}
	nf3 := c.root.children[0].children[0].children[0]
	if nf3.comment != "Attacks e5." || len(nf3.shapes) != 2 || nf3.shapes[1] != (boardShape{chess.B8, chess.C6, 1}) {
		t.Errorf("Nf3 comment %q shapes %v", nf3.comment, nf3.shapes)
	}
	// 2... d6 is a variation of 2... Nc6, with its own variation 3. Bc4
	if len(nf3.children) != 2 || nf3.children[1].san != "d6" {
		t.Fatalf("variations after Nf3: %v", nf3.children)
	}
	d6 := nf3.children[1]
	if d6.comment != "Philidor" || len(d6.children) != 2 {
		t.Fatalf("d6: comment %q, %d continuations (want 3. d4 and 3. Bc4)", d6.comment, len(d6.children))
	}
	if d4, alt := d6.children[0], d6.children[1]; d4.san != "d4" || alt.san != "Bc4" || d4.children[0].san != "Nf6" {
		t.Errorf("after 2... d6: %q (then %q), alternative %q", d4.san, d4.children[0].san, alt.san)
	}
	bc4 := nf3.children[0].children[0]
	if bc4.san != "Bc4" || plainTags(nagSuffix(bc4, "")) != "!? ⩲" {
		t.Errorf("Bc4 %q nags %q", bc4.san, plainTags(nagSuffix(bc4, "")))
	}
	if alt := bc4.children[1]; alt.san != "Nf6" || plainTags(nagSuffix(alt, "")) != "?!" || alt.children[0].san != "Ng5" {
		t.Errorf("3... Nf6?! line wrong")
	}

	// chapter from a position
	e := chs[1]
	if e.name() != "Endgame" || !e.white || e.problem != "" {
		t.Errorf("endgame: %q white %v problem %q", e.name(), e.white, e.problem)
	}
	if line := e.root.line(); len(line) != 5 || line[3].san != "e3+" {
		t.Errorf("endgame line: %d moves", len(line)-1)
	}

	// a move that can't be played: the chapter keeps what came before
	b := chs[2]
	if !strings.Contains(b.problem, `"Ke3"`) || len(b.root.line()) != 3 {
		t.Errorf("broken chapter: problem %q, %d moves", b.problem, len(b.root.line())-1)
	}
}

func TestStudyMoveText(t *testing.T) {
	c := parseStudyPGN(testStudyPGN)[0]
	text := studyMoveText(c, nil)
	for _, want := range []string{
		"The Italian Game.",
		`["n1"]e4[""]`,
		"Attacks e5.",
		`(`, "Philidor",
		`["n5"]d6[""]`,
		"3...",
		`Bc4[#ea45d8]!?[-] ⩲[""]`, // "interesting", in Lichess's pink
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "%cal") || strings.Contains(text, "%csl") {
		t.Error("drawing commands left in the text")
	}
	t.Log("\n" + text)
}

func TestEscapeText(t *testing.T) {
	for in, want := range map[string]string{
		"[11...dxe5 12. dxe5 Nd7 13. c4 holds the advantage]": "[11...dxe5 12. dxe5 Nd7 13. c4 holds the advantage]",
		"[red]":        "[red[]",
		"[#ff0000]":    "[#ff0000[]",
		`["region"]`:   `["region"[]`,
		"plain (text)": "plain (text)",
	} {
		if got := escapeText(in); got != want {
			t.Errorf("escapeText(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestStudyCommentEscapes(t *testing.T) {
	c := parseStudyPGN(`1. e4 { see [this] [%clk 0:01:00] } *`)[0]
	if c.root.children[0].comment != "see [this]" {
		t.Errorf("comment = %q", c.root.children[0].comment)
	}
	if text := studyMoveText(c, nil); !strings.Contains(text, "see [this[]") {
		t.Errorf("comment brackets not escaped:\n%s", text)
	}
}

func TestStudyName(t *testing.T) {
	c := parseStudyPGN(`[Event "My 60 Memorable Games: Fischer - Sherwin: Too little, too late"]
[ChapterName "Fischer - Sherwin: Too little, too late"]

1. e4 *`)[0]
	if c.studyName() != "My 60 Memorable Games" || c.name() != "Fischer - Sherwin: Too little, too late" {
		t.Errorf("study %q, chapter %q", c.studyName(), c.name())
	}
}

var colourTag = regexp.MustCompile(`\[(#[0-9a-f]{6}|-)\]`)

// plainTags drops colour tags.
func plainTags(s string) string { return colourTag.ReplaceAllString(s, "") }

func TestGlyphs(t *testing.T) {
	if glyphOf([]int{14, 4, 1}) != 4 || glyphOf([]int{14, 138}) != 14 || glyphOf([]int{999}) != 0 {
		t.Error("glyphOf: the first move symbol, else the first other symbol")
	}
	for _, g := range boardGlyphs {
		img := glyphImage(g, 40)
		c := img.RGBAAt(20, 3) // the disc's top edge: the glyph's colour
		want := g.color()
		if c.R != want.R || c.G != want.G || c.B != want.B {
			t.Errorf("%s badge colour %v, want %v", g.text(), c, want)
		}
		white := 0
		for y := 0; y < 40; y++ {
			for x := 0; x < 40; x++ {
				if p := img.RGBAAt(x, y); p.R > 200 && p.G > 200 && p.B > 200 {
					white++
				}
			}
		}
		if white < 20 {
			t.Errorf("%s badge has no symbol drawn (%d white pixels)", g.text(), white)
		}
	}
}

// The current move's comment and its side lines get different tints.
func TestStudyTints(t *testing.T) {
	c := parseStudyPGN(testStudyPGN)[0]
	atNf3 := studyMoveText(c, c.nodes[3]) // 2. Nf3 has a comment
	if !strings.Contains(atNf3, tintComment+"  Attacks e5.") && !strings.Contains(atNf3, tintComment+pgnComment+"Attacks e5.") {
		t.Errorf("comment not tinted:\n%s", atNf3)
	}
	if strings.Contains(atNf3, "[::u") {
		t.Error("no side lines branch at 2. Nf3, but some are tinted")
	}
	atNc6 := studyMoveText(c, c.nodes[4]) // 2... d6 branches off at 2... Nc6
	if !strings.Contains(atNc6, "[::u]"+pgnVariation+"(") || strings.Contains(atNc6, tintComment) {
		t.Errorf("side lines not tinted (or a comment wrongly is):\n%s", atNc6)
	}
}

func TestUnderlineLevels(t *testing.T) {
	// a side line at depth 1 with lines nested 1-5 levels deeper
	text := "  (a\n    (b)\n      (c)\n        (d)\n          (e)\n            (f)\n    " + pgnComment + "note)\n"
	got := strings.Split(underlineBranch(text, 1), "\n")
	want := []string{"[::u](a", "[::ul](b)", "[::uls](c)", "[::uli](d)", "[::ulis](e)", "[::ulis](f)", "[::u]" + pgnComment + "note)"}
	for i, w := range want {
		if !strings.Contains(got[i], w) {
			t.Errorf("line %d = %q, want %q", i, got[i], w)
		}
	}
}
