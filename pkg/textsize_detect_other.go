//go:build !unix

package pkg

// probeTerminal: the terminals with kitty's protocols don't run on Windows,
// so don't probe. CHESS_CLI_BIG_PIECES=1 can still force big pieces on.
func probeTerminal() (textSizing, graphics bool) { return false, false }
