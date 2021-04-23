from core import run_loop, Entity, style
def game(state, press):
  mut_state = state
  if press == "q":
    mut_state["running"] = False
  if press == "w":
    mut_state["entities"]["player"].y -= 1
  if press == "d":
    mut_state["entities"]["player"].x += 1
  if press == "s":
    mut_state["entities"]["player"].y += 1
  if press == "a":
    mut_state["entities"]["player"].x -= 1
  return mut_state
run_loop(game, {"running": True, "entities": {
  "player": Entity(10, 10, "@", style.RED)
  }})