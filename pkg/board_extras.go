package pkg

import (
	"fmt"
	"image"
	"image/color"
	"math"

	cv "codeberg.org/tslocum/cview"
	tc "github.com/gdamore/tcell/v3"
	"github.com/notnil/chess"
)

// Two Lichess board features: stepping back through the game, and drawing
// circles and arrows.
//
// History: the mouse wheel on the board (or [ and ], Home and End) shows
// earlier positions. Any click goes back to the current position, and a new
// move does too.
//
// Drawing: right-click a square for a circle, right-drag for an arrow; with
// Shift red, Alt blue, Shift+Alt yellow (green otherwise). Drawing the same
// shape again removes it; a left click clears them all. Shapes are cleared
// when the position changes.

// boardShape is a circle (from == to) or an arrow.
type boardShape struct {
	from, to chess.Square
	color    int // index into shapeColors
}

// Lichess's brush colours.
var shapeColors = []color.NRGBA{
	{0x15, 0x78, 0x1b, 0xff}, // green
	{0x88, 0x20, 0x20, 0xff}, // red
	{0x00, 0x30, 0x88, 0xff}, // blue
	{0xe6, 0x8f, 0x00, 0xff}, // yellow
}

func shapeColorFor(mod tc.ModMask) int {
	shift, alt := mod&tc.ModShift != 0, mod&tc.ModAlt != 0
	switch {
	case shift && alt:
		return 3
	case shift:
		return 1
	case alt:
		return 2
	}
	return 0
}

// SetHistory gives the board the game so far, for stepping back through it.
// positions has one more entry than moves (the start position).
func (b *BoardView) SetHistory(positions []*chess.Position, moves []*chess.Move) {
	b.mu.Lock()
	newMove := b.sounds && len(b.history) > 0 && len(positions) == len(b.history)+1 && len(moves) > 0
	if len(positions) != len(b.history) {
		b.back = 0 // a move was made: back to the present
	}
	b.history, b.historyMoves = positions, moves
	b.mu.Unlock()
	if newMove {
		playSound(moveSound(moves[len(moves)-1]))
	}
}

// EnableSounds makes a new move on this board play its sound.
func (b *BoardView) EnableSounds() { b.mu.Lock(); b.sounds = true; b.mu.Unlock() }

// ViewingHistory reports whether an earlier position is showing.
func (b *BoardView) ViewingHistory() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.back > 0
}

// step moves through history: positive = further back. Locked by caller.
func (b *BoardView) step(n int) {
	b.back = max(0, min(len(b.history)-1, b.back+n))
	if b.back > 0 {
		b.selected, b.targets, b.promo = chess.NoSquare, nil, nil
	}
}

// shown is the position to draw and its last move: the current one, or an
// earlier one while stepping back. Locked by caller.
func (b *BoardView) shown() (pos *chess.Position, from, to chess.Square) {
	if b.back == 0 || len(b.history) == 0 {
		return b.pos, b.lastFrom, b.lastTo
	}
	i := len(b.history) - 1 - b.back
	from, to = chess.NoSquare, chess.NoSquare
	if i > 0 && i-1 < len(b.historyMoves) {
		m := b.historyMoves[i-1]
		from, to = m.S1(), m.S2()
	}
	return b.history[i], from, to
}

// ShowPly shows the position after ply half-moves (0 = the start).
func (b *BoardView) ShowPly(ply int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.back = 0
	b.step(len(b.history) - 1 - ply)
}

// linkMoveList connects a move list (moveListText) to the board: clicking a
// move shows the position after it, and the move the board is showing is
// highlighted (following the wheel and [ ] too).
func linkMoveList(tv *cv.TextView, b *BoardView) {
	tv.SetRegions(true)
	moveLists = append(moveLists, tv)
	applyUIColours()
	syncing := false
	tv.SetHighlightedFunc(func(added, _, _ []string) {
		var ply int
		if syncing || len(added) == 0 {
			return
		}
		if _, err := fmt.Sscanf(added[0], "m%d", &ply); err != nil {
			return
		}
		b.ShowPly(ply)
		// the click focused the list; keep the focus where it was
		if prev := Root.App.GetFocus(); prev != nil && prev != tv {
			go Root.App.QueueUpdateDraw(func() { Root.App.SetFocus(prev) })
		}
	})
	b.mu.Lock()
	defer b.mu.Unlock()
	b.shownPly = -1
	b.onShown = func(ply int) { // UI goroutine
		syncing = true
		defer func() { syncing = false }()
		if ply <= 0 {
			tv.Highlight()
			return
		}
		tv.Highlight(fmt.Sprintf("m%d", ply))
		tv.ScrollToHighlight()
	}
}

// notifyShown tells the move list which move is showing, if it changed.
// Locked by caller (Draw).
func (b *BoardView) notifyShown() {
	if b.onShown == nil {
		return
	}
	ply := len(b.history) - 1 - b.back
	if ply == b.shownPly {
		return
	}
	b.shownPly = ply
	f := b.onShown
	go Root.App.QueueUpdateDraw(func() { f(ply) })
}

// historyLabel describes the position shown while stepping back, in at
// most width cells if it can (shorter forms for a narrow board or a clock
// sharing the row).
func (b *BoardView) historyLabel(width int) string {
	i, n := len(b.history)-1-b.back, len(b.history)-1
	for _, l := range []string{
		fmt.Sprintf(" move %d of %d · scroll down or ] to return ", i, n),
		fmt.Sprintf(" move %d of %d · ] to return ", i, n),
		fmt.Sprintf(" move %d/%d · ] returns ", i, n),
	} {
		if len([]rune(l)) <= width {
			return l
		}
	}
	return fmt.Sprintf(" %d/%d ", i, n)
}

// toggleShape adds a shape, or removes it if it's already drawn (in any
// colour). Locked by caller.
func (b *BoardView) toggleShape(s boardShape) {
	for i, x := range b.shapes {
		if x.from == s.from && x.to == s.to {
			b.shapes = append(b.shapes[:i], b.shapes[i+1:]...)
			if x.color == s.color {
				return
			}
			break
		}
	}
	b.shapes = append(b.shapes, s)
}

// SetShapes replaces the circles and arrows drawn on the board.
func (b *BoardView) SetShapes(shapes []boardShape) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.shapes = append([]boardShape(nil), shapes...)
}

// shapeTint is the square tint for shapes in text modes (where arrows can't
// be drawn): the circled square, or an arrow's ends.
func (b *BoardView) shapeTint(sq chess.Square) (tc.Color, bool) {
	for i := len(b.shapes) - 1; i >= 0; i-- {
		s := b.shapes[i]
		if s.from == sq || s.to == sq {
			c := shapeColors[s.color]
			return tc.NewRGBColor(int32(c.R), int32(c.G), int32(c.B)), true
		}
	}
	return 0, false
}

// shapesImage draws the shapes over the board, Lichess style, as one
// transparent image of the board's size in pixels.
func shapesImage(shapes []boardShape, white bool, sqPxW, sqPxH int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 8*sqPxW, 8*sqPxH))
	unit := float64(min(sqPxW, sqPxH))
	centre := func(sq chess.Square) (float64, float64) {
		row, col := displayOf(sq, white)
		return (float64(col) + 0.5) * float64(sqPxW), (float64(row) + 0.5) * float64(sqPxH)
	}
	for _, s := range shapes {
		c := shapeColors[s.color]
		x1, y1 := centre(s.from)
		if s.from == s.to {
			r, w := unit*0.47, unit*0.07
			paint(img, c, 0.8, x1-r, y1-r, x1+r, y1+r, func(x, y float64) float64 {
				return w/2 - math.Abs(math.Hypot(x-x1, y-y1)-(r-w/2))
			})
			continue
		}
		x2, y2 := centre(s.to)
		dx, dy := x2-x1, y2-y1
		l := math.Hypot(dx, dy)
		ux, uy := dx/l, dy/l
		head, halfHead, halfShaft := unit*0.42, unit*0.26, unit*0.08
		// the shaft starts a little off the centre and stops at the head
		sx, sy := x1+ux*unit*0.25, y1+uy*unit*0.25
		ex, ey := x2-ux*head, y2-uy*head
		pad := head + 2
		paint(img, c, 0.8, math.Min(x1, x2)-pad, math.Min(y1, y2)-pad, math.Max(x1, x2)+pad, math.Max(y1, y2)+pad,
			func(x, y float64) float64 {
				// signed distance inside the shaft or the head (positive = inside)
				t := (x-sx)*ux + (y-sy)*uy // along the arrow, from the shaft start
				n := -(x-sx)*uy + (y-sy)*ux
				shaftLen := (ex-sx)*ux + (ey-sy)*uy
				shaft := math.Min(halfShaft-math.Abs(n), math.Min(t, shaftLen+1-t))
				// head: triangle from (ex,ey) widening back to the tip at (x2,y2)
				th := (x-ex)*ux + (y-ey)*uy
				tri := math.Min(th, head-th) // between base and tip
				if tri > -1 {
					width := halfHead * (1 - th/head)
					tri = math.Min(tri, (width-math.Abs(n))*head/math.Hypot(head, halfHead))
				}
				return math.Max(shaft, tri)
			})
	}
	return img
}

// paint fills the pixels where inside(x, y) > 0 within a box, blending
// colour c at the given opacity, antialiased over one pixel.
func paint(img *image.RGBA, c color.NRGBA, opacity, x0, y0, x1, y1 float64, inside func(x, y float64) float64) {
	b := img.Bounds()
	for y := max(b.Min.Y, int(y0)); y < min(b.Max.Y, int(math.Ceil(y1))); y++ {
		for x := max(b.Min.X, int(x0)); x < min(b.Max.X, int(math.Ceil(x1))); x++ {
			cover := math.Max(0, math.Min(1, inside(float64(x)+0.5, float64(y)+0.5)+0.5))
			if cover == 0 {
				continue
			}
			a := cover * opacity
			i := img.PixOffset(x, y)
			p := img.Pix[i : i+4 : i+4]
			// "over" compositing in premultiplied RGBA
			p[0] = uint8(float64(c.R)*a + float64(p[0])*(1-a))
			p[1] = uint8(float64(c.G)*a + float64(p[1])*(1-a))
			p[2] = uint8(float64(c.B)*a + float64(p[2])*(1-a))
			p[3] = uint8(255*a + float64(p[3])*(1-a))
		}
	}
}

// Clock bars: a bar above and below the board, beside each player, showing
// how much of their time is left. Remaining time is gold, time used greys
// out, drawn in braille dots (two dot columns per cell, so the bar moves in
// half-cell steps). The running clock's bar is the brighter gold.

// clockBar is one player's clock: the time (label, e.g. "1:29") and a bar
// of the time left (frac 0-1, or < 0 for no bar, e.g. correspondence).
type clockBar struct {
	label  string
	right  string // shown at the other end, e.g. the time control
	frac   float64
	active bool   // this clock is running
	title  string // a player's title (GM, IM...), before the label
	plain  bool   // a player line (studies), not a clock: the label in plain text
}

func (c clockBar) shown() bool { return c.label != "" || c.frac >= 0 }

// SetClockBars sets the clocks for the players at the top and bottom of the
// board. A zero clockBar hides one.
func (b *BoardView) SetClockBars(top, bottom clockBar) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.topBar, b.bottomBar = top, bottom
}

// HideClockBars removes the clocks.
func (b *BoardView) HideClockBars() { b.SetClockBars(clockBar{frac: -1}, clockBar{frac: -1}) }

func (b *BoardView) hasClockBars() bool { return b.topBar.shown() || b.bottomBar.shown() }

// clockLabel formats seconds as m:ss (or h:mm:ss).
func clockLabel(s int) string {
	s = max(0, s)
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

var (
	colBarActive = tc.PaletteColor(220)     // gold
	colBarIdle   = tc.PaletteColor(178)     // darker gold
	colBarUsed   = tc.PaletteColor(240)     // grey
	colTitle     = tc.NewHexColor(0xbf811d) // Lichess's colour for titles (GM, IM...)
)

// drawClockBar draws a clock width cells wide at x, y: the time, then the bar.
func drawClockBar(screen tc.Screen, x, y, width int, bar clockBar, basic bool, bg tc.Color) {
	if !bar.shown() || width <= 0 {
		return
	}
	gold, used := colBarIdle, colBarUsed
	if bar.active {
		gold = colBarActive
	}
	if basic {
		gold, used = tc.ColorOlive, tc.ColorGray
		if bar.active {
			gold = tc.ColorYellow
		}
	}
	if bar.plain { // a player: title in Lichess's orange, then name and rating
		x0 := x
		put := func(s string, style tc.Style) {
			for _, r := range s {
				if x < x0+width {
					screen.SetContent(x, y, r, nil, style)
					x++
				}
			}
		}
		title, text := colTitle, tc.PaletteColor(252)
		if basic {
			title, text = tc.ColorOlive, tc.ColorSilver
		}
		if bar.title != "" {
			put(bar.title+" ", tc.StyleDefault.Foreground(title).Background(bg).Bold(true))
		}
		put(bar.label, tc.StyleDefault.Foreground(text).Background(bg))
		if right := []rune(bar.right); len(right) > 0 && x+1+len(right) <= x0+width {
			for i, r := range right {
				screen.SetContent(x0+width-len(right)+i, y, r, nil, tc.StyleDefault.Foreground(text).Background(bg).Bold(true))
			}
		}
		return
	}
	textCol := tc.PaletteColor(250)
	if basic {
		textCol = tc.ColorSilver
	}
	if bar.label != "" {
		// the running clock's time is highlighted, like on Lichess
		style := tc.StyleDefault.Foreground(textCol).Background(bg)
		if bar.active {
			style = tc.StyleDefault.Foreground(tc.ColorBlack).Background(gold).Bold(true)
		}
		label := []rune(" " + bar.label + " ")
		for i, r := range label {
			if i < width {
				screen.SetContent(x+i, y, r, nil, style)
			}
		}
		x += len(label) + 1
		width -= len(label) + 1
	}
	if bar.right != "" {
		right := []rune(" " + bar.right)
		if len(right) < width {
			for i, r := range right {
				screen.SetContent(x+width-len(right)+i, y, r, nil, tc.StyleDefault.Foreground(textCol).Background(bg))
			}
			width -= len(right) + 1
		}
	}
	if bar.frac < 0 || width <= 0 {
		return
	}
	units := int(math.Round(math.Max(0, math.Min(1, bar.frac)) * float64(2*width)))
	for i := 0; i < width; i++ {
		r, fg := '⣿', gold
		switch {
		case 2*i+2 <= units: // both dot columns: time left
		case 2*i+1 == units: // left column only
			r = '⡇'
		default: // time used
			fg = used
		}
		screen.SetContent(x+i, y, r, nil, tc.StyleDefault.Foreground(fg).Background(bg))
	}
}

// A king in check glows red, as on Lichess: a radial gradient behind it with
// image rendering, a red square otherwise.

var colCheck = tc.NewRGBColor(0xe7, 0, 0)

// shownMove is the move that led to the position shown, or nil (the start,
// or a board without its history). Locked by caller.
func (b *BoardView) shownMove() *chess.Move {
	if len(b.history) == 0 || len(b.historyMoves) != len(b.history)-1 {
		return nil
	}
	i := len(b.history) - 1 - b.back // index of the position shown
	if i < 1 {
		return nil
	}
	return b.historyMoves[i-1]
}

// checkSquare is the square of the king in check in the position shown
// (checkmate too), or chess.NoSquare. Locked by caller.
func (b *BoardView) checkSquare() chess.Square {
	m := b.shownMove()
	if m == nil || !m.HasTag(chess.Check) {
		return chess.NoSquare
	}
	pos, _, _ := b.shown()
	if pos == nil {
		return chess.NoSquare
	}
	king := chess.WhiteKing
	if pos.Turn() == chess.Black {
		king = chess.BlackKing
	}
	for sq, p := range pos.Board().SquareMap() {
		if p == king {
			return sq
		}
	}
	return chess.NoSquare
}

// checkGlowImage is the glow, size pixels square: Lichess's gradient, red at
// the centre fading out by the edge of the square (radial-gradient(ellipse
// at center, red 0%, #e70000 25%, rgba(169,0,0,0) 89%)).
func checkGlowImage(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	c := float64(size) / 2
	r := c * math.Sqrt2 // to the corners, as CSS's default
	type stop struct{ at, r, g, b, a float64 }
	stops := []stop{{0, 255, 0, 0, 1}, {0.25, 231, 0, 0, 1}, {0.89, 169, 0, 0, 0}, {1, 158, 0, 0, 0}}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			d := math.Hypot(float64(x)+0.5-c, float64(y)+0.5-c) / r
			s0, s1 := stops[len(stops)-2], stops[len(stops)-1]
			for i := 1; i < len(stops); i++ {
				if d <= stops[i].at {
					s0, s1 = stops[i-1], stops[i]
					break
				}
			}
			t := math.Max(0, math.Min(1, (d-s0.at)/(s1.at-s0.at)))
			a := s0.a + (s1.a-s0.a)*t
			lerp := func(u, v float64) uint8 { return uint8((u + (v-u)*t) * a) } // premultiplied
			img.SetRGBA(x, y, color.RGBA{lerp(s0.r, s1.r), lerp(s0.g, s1.g), lerp(s0.b, s1.b), uint8(255 * a)})
		}
	}
	return img
}
