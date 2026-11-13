//go:build !unix

package pkg

// detectTextSizing: kitty (the terminal with text sizing) doesn't run on
// Windows, so don't probe. CHESS_CLI_BIG_PIECES=1 can still force it on.
func detectTextSizing() bool { return false }
