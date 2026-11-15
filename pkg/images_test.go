package pkg

import (
	"strings"
	"testing"

	"github.com/notnil/chess"
)

// The image layer uploads pieces once per size and set, then only sends the
// placements that changed.
func TestImageLayerCommands(t *testing.T) {
	var il imageLayer
	place := func(ps ...imgPlacement) string {
		il.frame, il.frameSize, il.frameSet = ps, 40, "cburnett"
		return il.commands(100, 50)
	}
	wk := imgPlacement{x: 10, y: 5, img: pieceImageID(chess.WhiteKing), z: zPiece}
	bk := imgPlacement{x: 20, y: 5, img: pieceImageID(chess.BlackKing), z: zPiece}

	first := place(wk, bk)
	if !strings.Contains(first, "a=t,f=32,o=z") {
		t.Error("first frame didn't upload the pieces")
	}
	if strings.Count(first, "a=p,") != 2 {
		t.Errorf("first frame placements: %d", strings.Count(first, "a=p,"))
	}
	if !strings.HasPrefix(first, "\x1b7") || !strings.HasSuffix(first, "\x1b8") {
		t.Error("cursor not saved/restored around the images")
	}

	// same frame again: nothing to send
	if out := place(wk, bk); strings.Contains(out, "a=p,") || strings.Contains(out, "a=t,") {
		t.Errorf("unchanged frame sent %q", out)
	}

	// one piece moves: one placement, no upload
	moved := wk
	moved.x = 11
	out := place(moved, bk)
	if strings.Count(out, "a=p,") != 1 || strings.Contains(out, "a=t,") {
		t.Errorf("one move sent %q", out)
	}

	// a piece disappears: its placement is deleted
	out = place(moved)
	if !strings.Contains(out, "a=d,d=i,i=") || strings.Contains(out, "a=p,") {
		t.Errorf("removal sent %q", out)
	}

	// a different image at the same placement index: old one deleted first
	out = place(imgPlacement{x: 11, y: 5, img: pieceImageID(chess.WhiteQueen), z: zPiece})
	if !strings.Contains(out, "a=d,d=i,") || strings.Count(out, "a=p,") != 1 {
		t.Errorf("image swap sent %q", out)
	}

	// new piece set: everything uploaded again
	il.frame, il.frameSize, il.frameSet = []imgPlacement{wk}, 40, "merida"
	if out := il.commands(100, 50); !strings.Contains(out, "a=t,") {
		t.Error("piece set change didn't re-upload")
	}

	// board gone: all placements removed, then nothing more
	if out := place(); !strings.Contains(out, "a=d,d=i,") {
		t.Errorf("leaving the board sent %q", out)
	}
	if out := place(); out != "" {
		t.Errorf("idle sent %q", out)
	}
}

func TestPieceImagesExist(t *testing.T) {
	for _, set := range pieceSets {
		if n := len(loadPieceSet(set)); n != 12 {
			t.Errorf("piece set %s: %d of 12 pieces", set, n)
		}
	}
}
