from core import run_loop, Entity, style

pieces = {
    'R': '♜',
    'N': '♞',
    'B': '♗',
    'Q': '♕',
    'K': '♔',
    'P': '♙',
    'r': '♜',
    'n': '♞',
    'b': '♝',
    'q': '♛',
    'k': '♚',
    'p': '♙',

}
def game(state, press):
  mut_state = state
  if press == "q":
    mut_state["running"] = False

  return mut_state
run_loop(game, {"running": True, "entities": {
  "Rook": Entity(40, 40, pieces['R'], style.RED),
  "Knight": Entity(45, 40, pieces['N'], style.RED),
  }})
