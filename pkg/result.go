package pkg

import (
	"image"
	"image/color"
	"math"
	"strings"

	tc "github.com/gdamore/tcell/v3"
	"github.com/notnil/chess"
)

// Game results on the board, as Lichess shows them: a badge on each king when
// the game is over. Green crown for the winner, red # for a checkmated king,
// red flag for a loss on resignation or time, grey = for a draw. In text
// modes the king squares are tinted and get a badge character.

type badge int

const (
	badgeNone badge = iota
	badgeWin
	badgeMate
	badgeLoss
	badgeDraw
)

// boardResult is a finished game's result: winner (chess.NoColor for a draw)
// and how it ended ("mate", "resign", "timeout", "stalemate", "draw", ...).
type boardResult struct {
	set    bool
	winner chess.Color
	method string
}

// SetResult marks the game shown as over. The marks go when the position
// changes (a new game).
func (b *BoardView) SetResult(winner chess.Color, method string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.result = boardResult{set: true, winner: winner, method: method}
}

// badgeFor is the badge for a king of colour c. Locked by caller.
func (b *BoardView) badgeFor(c chess.Color) badge {
	r := b.result
	switch {
	case !r.set || b.back > 0:
		return badgeNone
	case r.winner == chess.NoColor:
		return badgeDraw
	case r.winner == c:
		return badgeWin
	case r.method == "mate":
		return badgeMate
	}
	return badgeLoss
}

var badgeColors = map[badge]color.NRGBA{
	badgeWin:  {0x62, 0x99, 0x24, 0xff},
	badgeMate: {0xcc, 0x33, 0x33, 0xff},
	badgeLoss: {0xcc, 0x33, 0x33, 0xff},
	badgeDraw: {0x88, 0x88, 0x88, 0xff},
}

// badgeText is the character shown in text modes.
var badgeText = map[badge]rune{badgeWin: '✓', badgeMate: '#', badgeLoss: '⚑', badgeDraw: '½'}

func badgeTc(k badge) tc.Color {
	c := badgeColors[k]
	return tc.NewRGBColor(int32(c.R), int32(c.G), int32(c.B))
}

func badgeBasic(k badge) tc.Color {
	switch k {
	case badgeWin:
		return tc.ColorGreen
	case badgeDraw:
		return tc.ColorGray
	}
	return tc.ColorMaroon
}

// badgeImage draws a badge: a coloured disc with a white symbol.
func badgeImage(k badge, size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	s := float64(size)
	c := s / 2
	white := color.NRGBA{0xff, 0xff, 0xff, 0xff}
	paint(img, badgeColors[k], 1, 0, 0, s, s, func(x, y float64) float64 { return c - 0.5 - math.Hypot(x-c, y-c) })
	bar := func(x0, y0, x1, y1 float64) { // a white rectangle, in units of the badge size
		paint(img, white, 1, x0*s, y0*s, x1*s, y1*s, func(x, y float64) float64 {
			return math.Min(math.Min(x-x0*s, x1*s-x), math.Min(y-y0*s, y1*s-y))
		})
	}
	poly := func(pts ...[2]float64) { // a white convex polygon (clockwise, badge units)
		paint(img, white, 1, 0, 0, s, s, func(x, y float64) float64 { return convexInside(pts, x/s, y/s) * s })
	}
	switch k {
	case badgeMate: // #
		bar(0.36, 0.22, 0.44, 0.78)
		bar(0.56, 0.22, 0.64, 0.78)
		bar(0.22, 0.36, 0.78, 0.44)
		bar(0.22, 0.56, 0.78, 0.64)
	case badgeDraw: // =
		bar(0.25, 0.36, 0.75, 0.45)
		bar(0.25, 0.55, 0.75, 0.64)
	case badgeLoss: // flag
		bar(0.30, 0.22, 0.37, 0.80)
		poly([2]float64{0.37, 0.22}, [2]float64{0.75, 0.37}, [2]float64{0.37, 0.52})
	case badgeWin: // crown
		bar(0.26, 0.60, 0.74, 0.72)
		poly([2]float64{0.26, 0.60}, [2]float64{0.26, 0.30}, [2]float64{0.42, 0.60})
		poly([2]float64{0.38, 0.60}, [2]float64{0.50, 0.24}, [2]float64{0.62, 0.60})
		poly([2]float64{0.58, 0.60}, [2]float64{0.74, 0.30}, [2]float64{0.74, 0.60})
	}
	return img
}

// convexInside is the signed distance inside a convex polygon (positive =
// inside), for either winding.
func convexInside(pts [][2]float64, x, y float64) float64 {
	d := math.Inf(1)
	area := 0.0
	for i := range pts {
		a, b := pts[i], pts[(i+1)%len(pts)]
		area += a[0]*b[1] - b[0]*a[1]
	}
	sign := 1.0
	if area < 0 {
		sign = -1
	}
	for i := range pts {
		a, b := pts[i], pts[(i+1)%len(pts)]
		ex, ey := b[0]-a[0], b[1]-a[1]
		l := math.Hypot(ex, ey)
		dist := sign * ((x-a[0])*ey - (y-a[1])*ex) / l
		d = math.Min(d, -dist)
	}
	return d
}

// methodText describes how a game ended.
var methodText = map[string]string{
	"mate": "checkmate", "resign": "resignation", "timeout": "time", "outoftime": "time",
	"stalemate": "stalemate", "draw": "agreement", "repetition": "repetition",
	"fiftymoves": "the 50-move rule", "material": "insufficient material", "aborted": "abort",
}

// resultText is a result for the sidebar: a big score and a line saying how.
func resultText(winner chess.Color, method string) string {
	how := methodText[method]
	if how == "" {
		how = method
	}
	switch winner {
	case chess.White:
		return "[::b]1–0[::-]\nWhite wins by " + how
	case chess.Black:
		return "[::b]0–1[::-]\nBlack wins by " + how
	}
	if how == "agreement" {
		return "[::b]½–½[::-]\nDraw"
	}
	return "[::b]½–½[::-]\nDraw by " + how
}

// gameMethod maps the chess library's ending to a method name.
func gameMethod(g *chess.Game) string {
	switch g.Method() {
	case chess.Checkmate:
		return "mate"
	case chess.Resignation:
		return "resign"
	case chess.Stalemate:
		return "stalemate"
	case chess.ThreefoldRepetition, chess.FivefoldRepetition:
		return "repetition"
	case chess.FiftyMoveRule, chess.SeventyFiveMoveRule:
		return "fiftymoves"
	case chess.InsufficientMaterial:
		return "material"
	case chess.DrawOffer:
		return "draw"
	}
	return strings.ToLower(g.Method().String())
}

// winnerOf maps an outcome to the winning colour (NoColor for a draw).
func winnerOf(o chess.Outcome) chess.Color {
	switch o {
	case chess.WhiteWon:
		return chess.White
	case chess.BlackWon:
		return chess.Black
	}
	return chess.NoColor
}

// colorFromName maps "white"/"black" to a colour ("" = NoColor).
func colorFromName(s string) chess.Color {
	switch strings.ToLower(s) {
	case "white":
		return chess.White
	case "black":
		return chess.Black
	}
	return chess.NoColor
}
