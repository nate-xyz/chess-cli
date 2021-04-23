from core import run_loop, Entity, style

pieces = {
    'K': '♔',
    'Q': '♕',
    'R': '♖',
    'B': '♗',
    'N': '♘',
    'P': '♙',
    'k': '♚',
    'q': '♛',
    'r': '♜',
    'b': '♝',
    'n': '♞',
    'p': '♟︎',

}
def game(state, press):
  mut_state = state
  if press == "q":
    mut_state["running"] = False

  return mut_state
run_loop(game, {"running": True, "entities": {

  #White First Rank
  "WhiteARook": Entity(40, 40, pieces['r'], style.WHITE),
  "WhiteBKnight": Entity(42, 40, pieces['n'], style.WHITE),
  "WhiteCBishop": Entity(44, 40, pieces['b'], style.WHITE),
  "WhiteQueen": Entity(46, 40, pieces['q'], style.WHITE),
  "WhiteKing": Entity(48, 40, pieces['k'], style.WHITE),
  "WhiteFBishop": Entity(50, 40, pieces['b'], style.WHITE),
  "WhiteGKnight": Entity(52, 40, pieces['n'], style.WHITE),
  "WhiteHRook": Entity(54, 40, pieces['r'], style.WHITE),

  #White Pawn Rank
  "WhiteAPawn": Entity(40, 38, pieces['p'], style.WHITE),
  "WhiteBPawn": Entity(42, 38, pieces['p'], style.WHITE),
  "WhiteCPawn": Entity(44, 38, pieces['p'], style.WHITE),
  "WhiteDPawn": Entity(46, 38, pieces['p'], style.WHITE),
  "WhiteEPawn": Entity(48, 38, pieces['p'], style.WHITE),
  "WhiteFPawn": Entity(50, 38, pieces['p'], style.WHITE),
  "WhiteGPawn": Entity(52, 38, pieces['p'], style.WHITE),
  "WhiteHPawn": Entity(54, 38, pieces['p'], style.WHITE),

  #Black First Rank
  "BlackARook": Entity(40, 30, pieces['R'], style.BLACK),
  "BlackBKnight": Entity(42, 30, pieces['N'], style.BLACK),
  "BlackCBishop": Entity(44, 30, pieces['B'], style.BLACK),
  "BlackQueen": Entity(46, 30, pieces['Q'], style.BLACK),
  "BlackKing": Entity(48, 30, pieces['K'], style.BLACK),
  "BlackFBishop": Entity(50, 30, pieces['B'], style.BLACK),
  "BlackGKnight": Entity(52, 30, pieces['N'], style.BLACK),
  "BlackHRook": Entity(54, 30, pieces['R'], style.BLACK),

  #Black Pawn Rank
  "BlackAPawn": Entity(40, 32, pieces['P'], style.BLACK),
  "BlackBPawn": Entity(42, 32, pieces['P'], style.BLACK),
  "BlackCPawn": Entity(44, 32, pieces['P'], style.BLACK),
  "BlackDPawn": Entity(46, 32, pieces['P'], style.BLACK),
  "BlackEPawn": Entity(48, 32, pieces['P'], style.BLACK),
  "BlackFPawn": Entity(50, 32, pieces['P'], style.BLACK),
  "BlackGPawn": Entity(52, 32, pieces['P'], style.BLACK),
  "BlackHPawn": Entity(54, 32, pieces['P'], style.BLACK),


  }})
