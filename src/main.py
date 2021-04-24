from core import run_loop, Entity, style
import random

pieces = {
    'K': '♔ ',
    'Q': '♕ ',
    'R': '♖ ',
    'B': '♗ ',
    'N': '♘ ',
    'P': '♙ ',
    'k': '♚ ',
    'q': '♛ ',
    'r': '♜ ',
    'b': '♝ ',
    'n': '♞ ',
    'p': '♟︎ ',

}

#rng to decide color (for testing flip)
playerColor = random.randint(1,2)

def game(state, press):
  mut_state = state
  if press == "q":
    mut_state["running"] = False

  return mut_state
if playerColor == 1:
    run_loop(game, {"running": True, "entities": {


        "WhatColor": Entity(40, 42, "You Play White", style.WHITE),

        #White First Rank
        "WhiteARook": Entity(20, 40, pieces['r'], style.WHITE),
        "WhiteBKnight": Entity(22, 40, pieces['n'], style.WHITE),
        "WhiteCBishop": Entity(24, 40, pieces['b'], style.WHITE),
        "WhiteQueen": Entity(26, 40, pieces['q'], style.WHITE),
        "WhiteKing": Entity(28, 40, pieces['k'], style.WHITE),
        "WhiteFBishop": Entity(30, 40, pieces['b'], style.WHITE),
        "WhiteGKnight": Entity(32, 40, pieces['n'], style.WHITE),
        "WhiteHRook": Entity(34, 40, pieces['r'], style.WHITE),

        #White Pawn Rank
        "WhiteAPawn": Entity(20, 38, pieces['P'], style.WHITE),
        "WhiteBPawn": Entity(22, 38, pieces['P'], style.WHITE),
        "WhiteCPawn": Entity(24, 38, pieces['P'], style.WHITE),
        "WhiteDPawn": Entity(26, 38, pieces['P'], style.WHITE),
        "WhiteEPawn": Entity(28, 38, pieces['P'], style.WHITE),
        "WhiteFPawn": Entity(30, 38, pieces['P'], style.WHITE),
        "WhiteGPawn": Entity(32, 38, pieces['P'], style.WHITE),
        "WhiteHPawn": Entity(34, 38, pieces['P'], style.WHITE),

        #Black First Rank
        "BlackARook": Entity(20, 30, pieces['R'], style.BLACK),
        "BlackBKnight": Entity(22, 30, pieces['N'], style.BLACK),
        "BlackCBishop": Entity(24, 30, pieces['B'], style.BLACK),
        "BlackQueen": Entity(26, 30, pieces['Q'], style.BLACK),
        "BlackKing": Entity(28, 30, pieces['K'], style.BLACK),
        "BlackFBishop": Entity(30, 30, pieces['B'], style.BLACK),
        "BlackGKnight": Entity(32, 30, pieces['N'], style.BLACK),
        "BlackHRook": Entity(34, 30, pieces['R'], style.BLACK),

        #Black Pawn Rank
        "BlackAPawn": Entity(20, 32, pieces['P'], style.BLACK),
        "BlackBPawn": Entity(22, 32, pieces['P'], style.BLACK),
        "BlackCPawn": Entity(24, 32, pieces['P'], style.BLACK),
        "BlackDPawn": Entity(26, 32, pieces['P'], style.BLACK),
        "BlackEPawn": Entity(28, 32, pieces['P'], style.BLACK),
        "BlackFPawn": Entity(30, 32, pieces['P'], style.BLACK),
        "BlackGPawn": Entity(32, 32, pieces['P'], style.BLACK),
        "BlackHPawn": Entity(34, 32, pieces['P'], style.BLACK),
        }})

elif playerColor == 2:
    run_loop(game, {"running": True, "entities": {


        "WhatColor": Entity(40, 42, "You Play Black", style.WHITE),

        #Black First Rank
        "BlackARook": Entity(20, 40, pieces['R'], style.BLACK),
        "BlackBKnight": Entity(22, 40, pieces['N'], style.BLACK),
        "BlackCBishop": Entity(24, 40, pieces['B'], style.BLACK),
        "BlackQueen": Entity(26, 40, pieces['Q'], style.BLACK),
        "BlackKing": Entity(28, 40, pieces['K'], style.BLACK),
        "BlackFBishop": Entity(30, 40, pieces['B'], style.BLACK),
        "BlackGKnight": Entity(32, 40, pieces['N'], style.BLACK),
        "BlackHRook": Entity(34, 40, pieces['R'], style.BLACK),

        #Blazck Pawn Rank
        "BlackAPawn": Entity(20, 38, pieces['P'], style.BLACK),
        "BlackBPawn": Entity(22, 38, pieces['P'], style.BLACK),
        "BlackCPawn": Entity(24, 38, pieces['P'], style.BLACK),
        "BlackDPawn": Entity(26, 38, pieces['P'], style.BLACK),
        "BlackEPawn": Entity(28, 38, pieces['P'], style.BLACK),
        "BlackFPawn": Entity(30, 38, pieces['P'], style.BLACK),
        "BlackGPawn": Entity(32, 38, pieces['P'], style.BLACK),
        "BlackHPawn": Entity(34, 38, pieces['P'], style.BLACK),

        #White First Rank
        "WhiteARook": Entity(20, 30, pieces['r'], style.WHITE),
        "WhiteBKnight": Entity(22, 30, pieces['n'], style.WHITE),
        "WhiteCBishop": Entity(24, 30, pieces['b'], style.WHITE),
        "WhiteQueen": Entity(26, 30, pieces['q'], style.WHITE),
        "WhiteKing": Entity(28, 30, pieces['k'], style.WHITE),
        "WhiteFBishop": Entity(30, 30, pieces['b'], style.WHITE),
        "WhiteGKnight": Entity(32, 30, pieces['n'], style.WHITE),
        "WhiteHRook": Entity(34, 30, pieces['r'], style.WHITE),

        #White Pawn Rank
<<<<<<< HEAD
        "WhiteAPawn": Entity(20, 32, pieces['p'], style.WHITE),
        "WhiteBPawn": Entity(22, 32, pieces['p'], style.WHITE),
        "WhiteCPawn": Entity(24, 32, pieces['p'], style.WHITE),
        "WhiteDPawn": Entity(26, 32, pieces['p'], style.WHITE),
        "WhiteEPawn": Entity(28, 32, pieces['p'], style.WHITE),
        "WhiteFPawn": Entity(30, 32, pieces['p'], style.WHITE),
        "WhiteGPawn": Entity(32, 32, pieces['p'], style.WHITE),
        "WhiteHPawn": Entity(34, 32, pieces['p'], style.WHITE),
=======
        "WhiteAPawn": Entity(40, 32, pieces['P'], style.WHITE),
        "WhiteBPawn": Entity(42, 32, pieces['P'], style.WHITE),
        "WhiteCPawn": Entity(44, 32, pieces['P'], style.WHITE),
        "WhiteDPawn": Entity(46, 32, pieces['P'], style.WHITE),
        "WhiteEPawn": Entity(48, 32, pieces['P'], style.WHITE),
        "WhiteFPawn": Entity(50, 32, pieces['P'], style.WHITE),
        "WhiteGPawn": Entity(52, 32, pieces['P'], style.WHITE),
        "WhiteHPawn": Entity(54, 32, pieces['P'], style.WHITE),
>>>>>>> 6c7e8f7891a0444cd564e80c26cd3d43f02a2af9
        }})
