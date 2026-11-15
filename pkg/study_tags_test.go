package pkg

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
)

const taggedPGN = `[Event "1972 World Chess Championship: Game 1"]
[Site "Reykjavík ISL"]
[Date "1972.07.11"]
[Round "1"]
[White "Boris Vasilyevich Spassky"]
[Black "Robert James Fischer"]
[WhiteTitle "GM"]
[BlackTitle "GM"]
[WhiteElo "2660"]
[BlackElo "2785"]
[Result "1-0"]
[TimeControl "40/9000:3600"]
[ECO "E56"]
[Opening "Nimzo-Indian Defense: Normal Variation"]
[Annotator "https://lichess.org/@/someone"]
[ChapterName "Game 1"]

1. d4 Nf6 1-0`

func TestChapterTags(t *testing.T) {
	c := parseStudyPGN(taggedPGN)[0]
	want := "Reykjavík ISL · 11 Jul 1972 · Round 1 · 150 min for 40 moves, then 60 min · E56 Nimzo-Indian Defense: Normal Variation"
	if got := chapterTags(c); got != want {
		t.Errorf("tags:\n got %q\nwant %q", got, want)
	}
	w, b := playerBar(c, chess.White), playerBar(c, chess.Black)
	if w.title != "GM" || w.label != "Boris Vasilyevich Spassky 2660" || w.right != "1" || !w.plain {
		t.Errorf("white: %+v", w)
	}
	if b.label != "Robert James Fischer 2785" || b.right != "0" {
		t.Errorf("black: %+v", b)
	}
	if bar := playerBar(parseStudyPGN(`1. e4 *`)[0], chess.White); bar.shown() {
		t.Error("no player, but a player line")
	}
}

func TestPGNDatesAndClocks(t *testing.T) {
	for in, want := range map[string]string{"1972.07.11": "11 Jul 1972", "1972.07.??": "Jul 1972", "1972.??.??": "1972", "oops": "oops"} {
		if got := pgnDate(in); got != want {
			t.Errorf("pgnDate(%q) = %q, want %q", in, got, want)
		}
	}
	for in, want := range map[string]string{"5400+30": "90+30", "180+2": "3+2", "600": "10 min", "40/7200:3600": "120 min for 40 moves, then 60 min", "weird": "weird"} {
		if got := pgnTimeControl(in); got != want {
			t.Errorf("pgnTimeControl(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestChapterResult(t *testing.T) {
	for pgn, want := range map[string]string{
		taggedPGN: "1–0 White is victorious",
		`[Result "0-1"]` + "\n\n1. f3 e5 2. g4 Qh4# 0-1":                           "0–1 Checkmate • Black is victorious",
		"1. f3 e5 2. g4 Qh4# *":                                                    "0–1 Checkmate • Black is victorious", // the position says so
		`[Result "1-0"]` + "\n" + `[Termination "Time forfeit"]` + "\n\n1. e4 1-0": "1–0 Time out • White is victorious",
		`[Result "1/2-1/2"]` + "\n\n1. e4 e5 1/2-1/2":                              "½–½ Draw",
		"1. e4 e5 *": "",
	} {
		score, how := chapterResult(parseStudyPGN(pgn)[0])
		if got := strings.TrimSpace(score + " " + how); got != want {
			t.Errorf("%q: got %q, want %q", pgn, got, want)
		}
	}
	if text := studyMoveText(parseStudyPGN(taggedPGN)[0], nil); !strings.Contains(text, "[::b]1–0[::-]") {
		t.Errorf("result missing from the move list:\n%s", text)
	}
}
