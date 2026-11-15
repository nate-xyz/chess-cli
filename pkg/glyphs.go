package pkg

import (
	"image"
	"image/color"
	"math"
	"sync"

	tc "github.com/gdamore/tcell/v3"
	"github.com/notnil/chess"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// Annotation symbols on the board, as Lichess shows them in studies: a
// coloured disc in the top-right corner of the square the move went to,
// with the move's symbol (!, ??, □...) or, failing that, its position or
// idea symbol (±, ⊕, →...).

// moveGlyph is a symbol, by its NAG number; 0 is none.
type moveGlyph int

// boardGlyphs are the symbols drawn on the board, move symbols first (the
// order they're preferred in), each with its own image.
var boardGlyphs = []moveGlyph{1, 2, 3, 4, 5, 6, 7, 22, 10, 13, 14, 15, 16, 17, 18, 19, 32, 36, 40, 44, 132, 138, 140, 146}

// Lichess's colours for move symbols; the others share one colour.
var glyphColors = map[moveGlyph]color.NRGBA{
	1: {0x22, 0xac, 0x38, 0xff}, // ! good
	2: {0xe6, 0x9d, 0x00, 0xff}, // ? mistake
	3: {0x16, 0x82, 0x26, 0xff}, // !! brilliant
	4: {0xdf, 0x53, 0x53, 0xff}, // ?? blunder
	5: {0xea, 0x45, 0xd8, 0xff}, // !? interesting
	6: {0x56, 0xb4, 0xe9, 0xff}, // ?! dubious
}

var glyphOther = color.NRGBA{0xa7, 0x08, 0x4e, 0xff} // ⊕, ±, → ... as on Lichess

func (g moveGlyph) color() color.NRGBA {
	if c, ok := glyphColors[g]; ok {
		return c
	}
	return glyphOther
}

// glyphBasic is the nearest of the 8 standard colours, for basic rendering.
func (g moveGlyph) basic() tc.Color {
	switch g {
	case 1, 3:
		return tc.ColorGreen
	case 2:
		return tc.ColorOlive
	case 4:
		return tc.ColorMaroon
	case 6:
		return tc.ColorNavy
	}
	return tc.ColorPurple
}

func (g moveGlyph) text() string { t, _ := nagText(int(g)); return t }

func (g moveGlyph) tc() tc.Color {
	c := g.color()
	return tc.NewRGBColor(int32(c.R), int32(c.G), int32(c.B))
}

// glyphOf is the symbol to show for a move: its first move symbol, or its
// first other symbol.
func glyphOf(nags []int) moveGlyph {
	for _, want := range []bool{true, false} {
		for _, n := range nags {
			if t, isMove := nagText(n); t != "" && isMove == want && glyphIndex(moveGlyph(n)) >= 0 {
				return moveGlyph(n)
			}
		}
	}
	return 0
}

func glyphIndex(g moveGlyph) int {
	for i, x := range boardGlyphs {
		if x == g {
			return i
		}
	}
	return -1
}

// SetGlyph shows a symbol on a square (g = 0 removes it).
func (b *BoardView) SetGlyph(sq chess.Square, g moveGlyph) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.glyphSq, b.glyph = sq, g
}

var glyphFont = sync.OnceValue(func() *opentype.Font {
	f, _ := opentype.Parse(gobold.TTF)
	return f
})

// glyphImage draws a symbol's badge: the coloured disc with the symbol in
// white, size pixels square.
func glyphImage(g moveGlyph, size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	s := float64(size)
	c := s / 2
	white := color.NRGBA{0xff, 0xff, 0xff, 0xff}
	paint(img, g.color(), 1, 0, 0, s, s, func(x, y float64) float64 { return c - 0.5 - math.Hypot(x-c, y-c) })
	ring := func(x, y float64) float64 { return 0.04*s - math.Abs(math.Hypot(x-c, y-c)-0.26*s) }
	switch t := g.text(); t {
	case "⊕": // the font has no circled plus or dot: drawn
		arm := 0.04 * s
		paint(img, white, 1, 0, 0, s, s, func(x, y float64) float64 {
			plus := math.Max(math.Min(arm-math.Abs(x-c), 0.26*s-math.Abs(y-c)), math.Min(arm-math.Abs(y-c), 0.26*s-math.Abs(x-c)))
			return math.Max(ring(x, y), plus)
		})
	case "□": // the font's square is tiny
		half, w := 0.19*s, 0.05*s
		paint(img, white, 1, 0, 0, s, s, func(x, y float64) float64 {
			d := math.Max(math.Abs(x-c), math.Abs(y-c))
			return w/2 - math.Abs(d-(half-w/2))
		})
	case "⨀":
		paint(img, white, 1, 0, 0, s, s, func(x, y float64) float64 {
			return math.Max(ring(x, y), 0.08*s-math.Hypot(x-c, y-c))
		})
	case "⩲": // nor these: two characters stacked
		stackText(img, "+", "=")
	case "⩱":
		stackText(img, "=", "+")
	case "∓":
		stackText(img, "−", "+")
	case "⇆":
		stackText(img, "←", "→")
	default:
		drawText(img, t, 0.62, c)
	}
	return img
}

// drawText draws white text centred horizontally, its ink centred on y.
// size is the font size as a fraction of the badge; wide text (+−, =∞) is
// shrunk to fit inside the disc.
func drawText(img *image.RGBA, text string, size, y float64) {
	s := float64(img.Bounds().Dx())
	face, err := opentype.NewFace(glyphFont(), &opentype.FaceOptions{Size: size * s, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		return
	}
	d := &font.Drawer{Dst: img, Src: image.White, Face: face}
	ink, _ := d.BoundString(text)
	w := float64(ink.Max.X-ink.Min.X) / 64
	if maxW := 0.6 * s; w > maxW {
		face.Close()
		face, err = opentype.NewFace(glyphFont(), &opentype.FaceOptions{Size: size * s * maxW / w, DPI: 72, Hinting: font.HintingNone})
		if err != nil {
			return
		}
		d.Face = face
		ink, _ = d.BoundString(text)
		w = float64(ink.Max.X-ink.Min.X) / 64
	}
	defer face.Close()
	h := float64(ink.Max.Y-ink.Min.Y) / 64
	x0 := (s-w)/2 - float64(ink.Min.X)/64
	y0 := y - h/2 - float64(ink.Min.Y)/64
	d.Dot = fixed.Point26_6{X: fixed.Int26_6(x0 * 64), Y: fixed.Int26_6(y0 * 64)}
	d.DrawString(text)
}

// stackText draws two characters one above the other.
func stackText(img *image.RGBA, top, bottom string) {
	s := float64(img.Bounds().Dx())
	drawText(img, top, 0.46, 0.35*s)
	drawText(img, bottom, 0.46, 0.67*s)
}
