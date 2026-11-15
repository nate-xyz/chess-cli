package pkg

import (
	"bytes"
	"compress/zlib"
	"embed"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math"
	"strings"
	"sync"

	tc "github.com/gdamore/tcell/v3"
	"github.com/nate-xyz/chess-cli/api"
	"github.com/notnil/chess"
	xdraw "golang.org/x/image/draw"
)

// Piece images via kitty's graphics protocol
// (https://sw.kovidgoyal.net/kitty/graphics-protocol/), supported by kitty,
// Ghostty and others.
//
// The board squares are still ordinary cells coloured by tcell; only the
// pieces (and the legal-move markers) are images, placed on top of the
// squares. Each piece is scaled to the square's size in pixels and uploaded
// once; moving a piece just moves a placement, so nothing is re-sent until
// the board changes size or the piece set changes. The piece sets come from
// Lichess (see assets/pieces/README.md for their licences).

//go:embed assets/pieces/*/*.png
var pieceFiles embed.FS

// pieceSets are the bundled piece sets, Lichess's default first.
var pieceSets = []string{"cburnett", "merida", "mpchess", "chessnut", "fantasy", "spatial",
	"celtic", "kiwen-suwi", "totoy", "papercut", "rhosgfx", "shapes"}

// boardTheme is a board's square colours (Lichess's flat-colour boards).
type boardTheme struct {
	name        string
	light, dark tc.Color
}

var boardThemes = []boardTheme{
	{"brown", tc.NewRGBColor(0xf0, 0xd9, 0xb5), tc.NewRGBColor(0xb5, 0x88, 0x63)},
	{"blue", tc.NewRGBColor(0xde, 0xe3, 0xe6), tc.NewRGBColor(0x8c, 0xa2, 0xad)},
	{"green", tc.NewRGBColor(0xff, 0xff, 0xdd), tc.NewRGBColor(0x86, 0xa6, 0x66)},
	{"purple", tc.NewRGBColor(0x9f, 0x90, 0xb0), tc.NewRGBColor(0x7d, 0x4a, 0x8d)},
	{"ic", tc.NewRGBColor(0xec, 0xec, 0xec), tc.NewRGBColor(0xc1, 0xc1, 0x8e)},
	{"grey", tc.NewRGBColor(0xdc, 0xdc, 0xdc), tc.NewRGBColor(0xab, 0xab, 0xab)},
}

// currentPieceSet and currentBoard are the chosen piece set and board.
func currentPieceSet() string {
	for _, s := range pieceSets {
		if s == api.UserInfo.PieceSet {
			return s
		}
	}
	return pieceSets[0]
}

func currentBoard() boardTheme {
	for _, t := range boardThemes {
		if t.name == api.UserInfo.Board {
			return t
		}
	}
	return boardThemes[0]
}

// Image ids: 1-12 the pieces, then the move dot and the capture ring. Offset
// so they're unlikely to clash with anything else using the protocol.
const (
	imgBase    = 0x5c4e00
	imgDot     = imgBase + 13
	imgRing    = imgBase + 14
	imgOverlay = imgBase + 15 // drawn circles and arrows
	imgBadges  = imgBase + 16 // result badges: +0 win, +1 mate, +2 loss, +3 draw
	imgGlyphs  = imgBase + 20 // annotation symbols, in boardGlyphs order
	imgCount   = 20 + 24 - 1

	zMarker = 1 // z-indexes: above the cell text, pieces over markers,
	zPiece  = 2 // shapes over pieces
	zShapes = 3
	zBadge  = 4
)

func imgBadge(k badge) int { return imgBadges + int(k) - int(badgeWin) }

func imgGlyph(g moveGlyph) int { return imgGlyphs + glyphIndex(g) }

// badgeSizeFor is a result badge's size for a piece size.
func badgeSizeFor(pieceSize int) int { return max(8, pieceSize*2/5) }

func pieceImageID(p chess.Piece) int {
	return imgBase + int(p) // chess.Piece: WhiteKing = 1 ... BlackPawn = 12
}

// imgPlacement is one image placed at cell (x, y), offset (px, py) pixels
// into that cell.
type imgPlacement struct {
	x, y, px, py int
	img, z       int
}

type imageLayer struct {
	sync.Mutex
	frame     []imgPlacement // collected by board Draw calls this frame
	frameSize int            // square size in pixels the frame's boards want
	frameSet  string         // piece set the frame's boards want

	overlay     *image.RGBA // the frame's drawn shapes, if any
	overlayKey  string      // identifies the overlay's content
	sentOverlay string      // key of the overlay on the terminal

	prev        []imgPlacement // placements on the terminal now
	uploaded    int            // size in pixels of the uploaded images (0 = none)
	uploadedSet string         // and their piece set
	prevW       int
	prevH       int

	sources    map[chess.Piece]image.Image
	sourcesSet string
}

var imgLayer imageLayer

// cellPixels is the size of one terminal cell in pixels, if known.
func cellPixels(screen tc.Screen) (w, h int, ok bool) {
	tty, ok := screen.Tty()
	if !ok {
		return 0, 0, false
	}
	ws, err := tty.WindowSize()
	if err != nil {
		return 0, 0, false
	}
	w, h = ws.CellDimensions()
	return w, h, w > 0 && h > 0
}

// queueImages records image placements for this frame. Called from Draw.
func queueImages(size int, set string, ps []imgPlacement) {
	imgLayer.Lock()
	defer imgLayer.Unlock()
	imgLayer.frameSize, imgLayer.frameSet = size, set
	imgLayer.frame = append(imgLayer.frame, ps...)
}

// queueOverlay records the board's drawn shapes for this frame.
func queueOverlay(img *image.RGBA, key string) {
	imgLayer.Lock()
	defer imgLayer.Unlock()
	imgLayer.overlay, imgLayer.overlayKey = img, key
}

// flushImages runs after each frame (after tcell has drawn): uploads the
// images if the size changed and updates the placements that changed.
func flushImages(screen tc.Screen) {
	tty, ok := screen.Tty()
	if !ok {
		return
	}
	w, h := screen.Size()
	imgLayer.Lock()
	out := imgLayer.commands(w, h)
	imgLayer.Unlock()
	if out != "" {
		_, _ = tty.Write([]byte(out))
	}
}

// commands works out the escape codes to bring the terminal's images up to
// date with this frame (for a w x h screen), and takes the frame. Locked.
func (il *imageLayer) commands(w, h int) string {
	frame, size, set := il.frame, il.frameSize, il.frameSet
	overlay, overlayKey := il.overlay, il.overlayKey
	il.frame, il.overlay, il.overlayKey = nil, nil, ""
	if len(frame) == 0 && len(il.prev) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\x1b7")
	if w != il.prevW || h != il.prevH {
		// after a resize tcell has cleared and redrawn the screen: start over
		deletePlacements(&b, il.prev, 0)
		il.prev = nil
		il.prevW, il.prevH = w, h
	}
	if len(frame) > 0 && (size != il.uploaded || set != il.uploadedSet) {
		for id := imgBase + 1; id <= imgBase+imgCount; id++ {
			fmt.Fprintf(&b, "\x1b_Ga=d,d=I,i=%d,q=2\x1b\\", id) // placements and data
		}
		il.prev = nil
		il.uploadImages(&b, size, set)
		il.uploaded, il.uploadedSet = size, set
		il.sentOverlay = ""
	}
	if overlay != nil && overlayKey != il.sentOverlay {
		// replace the shapes image; its placement has to be made again
		fmt.Fprintf(&b, "\x1b_Ga=d,d=I,i=%d,q=2\x1b\\", imgOverlay)
		transmit(&b, imgOverlay, overlay)
		il.sentOverlay = overlayKey
		for i := range il.prev {
			if il.prev[i].img == imgOverlay {
				il.prev[i].z = -1 // never equal: placed again below
			}
		}
	}
	for i, p := range frame {
		if i < len(il.prev) {
			if il.prev[i] == p {
				continue
			}
			if il.prev[i].img != p.img { // placement ids are per image
				deletePlacements(&b, il.prev[i:i+1], i)
			}
		}
		fmt.Fprintf(&b, "\x1b[%d;%dH\x1b_Ga=p,i=%d,p=%d,X=%d,Y=%d,z=%d,C=1,q=2\x1b\\",
			p.y+1, p.x+1, p.img, i+1, p.px, p.py, p.z)
	}
	if len(il.prev) > len(frame) {
		deletePlacements(&b, il.prev[len(frame):], len(frame))
	}
	il.prev = frame
	b.WriteString("\x1b8")
	return b.String()
}

// deletePlacements removes placements, the first of which has id first+1.
func deletePlacements(b *strings.Builder, ps []imgPlacement, first int) {
	for i, p := range ps {
		fmt.Fprintf(b, "\x1b_Ga=d,d=i,i=%d,p=%d,q=2\x1b\\", p.img, first+i+1)
	}
}

// uploadImages sends all the images at size x size pixels, as zlib-compressed
// RGBA in chunks (the protocol's limit is 4096 bytes of base64 per chunk).
func (il *imageLayer) uploadImages(b *strings.Builder, size int, set string) {
	if il.sources == nil || il.sourcesSet != set {
		il.sources, il.sourcesSet = loadPieceSet(set), set
	}
	for p := chess.WhiteKing; p <= chess.BlackPawn; p++ {
		if src, ok := il.sources[p]; ok {
			dst := image.NewRGBA(image.Rect(0, 0, size, size))
			xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
			transmit(b, pieceImageID(p), dst)
		}
	}
	transmit(b, imgDot, markerImage(size, false))
	transmit(b, imgRing, markerImage(size, true))
	for k := badgeWin; k <= badgeDraw; k++ {
		transmit(b, imgBadge(k), badgeImage(k, badgeSizeFor(size)))
	}
	for _, g := range boardGlyphs {
		transmit(b, imgGlyph(g), glyphImage(g, badgeSizeFor(size)))
	}
}

// loadPieceSet decodes a piece set's embedded images.
func loadPieceSet(set string) map[chess.Piece]image.Image {
	sources := map[chess.Piece]image.Image{}
	for p := chess.WhiteKing; p <= chess.BlackPawn; p++ {
		name := "assets/pieces/" + set + "/" + "wb"[p.Color()-1:p.Color()] + strings.ToUpper(p.Type().String()) + ".png"
		f, err := pieceFiles.Open(name)
		if err != nil {
			continue
		}
		if img, _, err := image.Decode(f); err == nil {
			sources[p] = img
		}
		f.Close()
	}
	return sources
}

// colMoveDest is Lichess's legal-move marker colour.
var colMoveDest = color.NRGBA{20, 85, 30, 128}

// markerImage draws Lichess's legal-move markers: a dot for an empty square,
// or for a capture a tint in the corners around a circle.
func markerImage(size int, ring bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	c := float64(size) / 2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			d := math.Hypot(float64(x)+0.5-c, float64(y)+0.5-c)
			var cover float64 // fraction of the pixel covered, for smooth edges
			if ring {
				cover = d - 0.56*float64(size) + 0.5
			} else {
				cover = 0.17*float64(size) - d + 0.5
			}
			cover = math.Max(0, math.Min(1, cover))
			if cover > 0 {
				col := colMoveDest
				col.A = uint8(float64(col.A) * cover)
				img.Set(x, y, col)
			}
		}
	}
	return img
}

// transmit uploads an image (a=t: store without showing) under an id.
func transmit(b *strings.Builder, id int, img *image.RGBA) {
	var z bytes.Buffer
	w := zlib.NewWriter(&z)
	_, _ = w.Write(img.Pix)
	_ = w.Close()
	data := base64.StdEncoding.EncodeToString(z.Bytes())
	sz := img.Bounds().Size()
	for first := true; first || len(data) > 0; first = false {
		chunk := data[:min(4096, len(data))]
		data = data[len(chunk):]
		more := 0
		if len(data) > 0 {
			more = 1
		}
		if first {
			fmt.Fprintf(b, "\x1b_Ga=t,f=32,o=z,i=%d,s=%d,v=%d,q=2,m=%d;%s\x1b\\", id, sz.X, sz.Y, more, chunk)
		} else {
			fmt.Fprintf(b, "\x1b_Gm=%d;%s\x1b\\", more, chunk)
		}
	}
}
