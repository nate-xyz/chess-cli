package pkg

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/notnil/chess"
)

// A study chapter's tags (PGN headers), shown as Lichess does: the game's
// details under the chapter title, and each player beside the board with
// their title, name, rating and result.

// tag is a header's value, or "" if it's missing or unknown ("?", "-").
func (c *studyChapter) tag(key string) string {
	v := strings.TrimSpace(c.headers[key])
	if v == "?" || v == "-" || v == "????.??.??" || v == "*" {
		return ""
	}
	return v
}

// chapterTags is the line of game details: where, when, round, time
// control, opening, annotator and variant, those that are known.
func chapterTags(c *studyChapter) string {
	var parts []string
	add := func(s string) {
		if s != "" {
			parts = append(parts, s)
		}
	}
	if site := c.tag("Site"); !strings.Contains(site, "lichess.org/") { // Lichess puts the study link there
		add(site)
	}
	add(pgnDate(c.tag("Date")))
	if r := c.tag("Round"); r != "" {
		add("Round " + r)
	}
	add(pgnTimeControl(c.tag("TimeControl")))
	opening := strings.TrimSpace(c.tag("ECO") + " " + c.tag("Opening"))
	add(opening)
	if a := c.tag("Annotator"); a != "" && !strings.Contains(a, "lichess.org/") {
		add("Annotated by " + a)
	}
	if v := c.tag("Variant"); v != "" && v != "Standard" && v != "From Position" {
		add(v)
	}
	return strings.Join(parts, " · ")
}

// pgnDate turns "1972.07.11" into "11 Jul 1972" (or "Jul 1972", "1972"
// when the day or month is unknown).
func pgnDate(d string) string {
	f := strings.Split(d, ".")
	if len(f) != 3 || strings.Contains(f[0], "?") {
		return d
	}
	if t, err := time.Parse("2006.01.02", d); err == nil {
		return t.Format("2 Jan 2006")
	}
	if t, err := time.Parse("2006.01", f[0]+"."+f[1]); err == nil {
		return t.Format("Jan 2006")
	}
	return f[0]
}

// pgnTimeControl reads a PGN TimeControl: "5400+30" is "90+30", and
// "40/7200:3600" is "120 min for 40 moves, then 60 min".
func pgnTimeControl(tcText string) string {
	if tcText == "" {
		return ""
	}
	mins := func(s string) (string, bool) {
		n, err := strconv.Atoi(s)
		if err != nil {
			return "", false
		}
		if n%60 == 0 {
			return strconv.Itoa(n / 60), true
		}
		return fmt.Sprintf("%.1f", float64(n)/60), true
	}
	var periods []string
	for _, p := range strings.Split(tcText, ":") {
		switch {
		case strings.Contains(p, "/"):
			moves, secs, _ := strings.Cut(p, "/")
			m, ok := mins(secs)
			if !ok {
				return tcText
			}
			periods = append(periods, fmt.Sprintf("%s min for %s moves", m, moves))
		case strings.Contains(p, "+"):
			base, inc, _ := strings.Cut(p, "+")
			m, ok := mins(base)
			if _, err := strconv.Atoi(inc); !ok || err != nil {
				return tcText
			}
			periods = append(periods, m+"+"+inc)
		default:
			m, ok := mins(p)
			if !ok {
				return tcText
			}
			periods = append(periods, m+" min")
		}
	}
	return strings.Join(periods, ", then ")
}

// playerBar is a player's line beside the board: title, name and rating,
// and their result at the other end. A zero bar (hidden) if there's no name.
func playerBar(c *studyChapter, side chess.Color) clockBar {
	key := "White"
	if side == chess.Black {
		key = "Black"
	}
	name := c.tag(key)
	if name == "" {
		return clockBar{frac: -1}
	}
	label := name
	if elo := c.tag(key + "Elo"); elo != "" && elo != "0" {
		label += " " + elo
	}
	bar := clockBar{title: c.tag(key + "Title"), label: label, frac: -1, plain: true}
	switch c.tag("Result") {
	case "1-0":
		bar.right = map[chess.Color]string{chess.White: "1", chess.Black: "0"}[side]
	case "0-1":
		bar.right = map[chess.Color]string{chess.White: "0", chess.Black: "1"}[side]
	case "1/2-1/2", "½-½":
		bar.right = "½"
	}
	return bar
}
