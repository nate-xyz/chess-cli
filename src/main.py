from core import run_loop, Entity, style
import random
import chess

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

piece_name = {
    'K': 'WhiteKing',
    'Q': 'WhiteQueen',
    'R': 'WhiteRook',
    'B': 'WhiteBishop',
    'N': 'WhiteKnight',
    'P': 'WhitePawn',
    'k': 'BlackKing',
    'q': 'BlackQueen',
    'r': 'BlackRook',
    'b': 'BlackBishop',
    'n': 'BlackKnight',
    'p': 'BlackPawn',
}

rank = {
    20: 'a',
    22: 'b',
    24: 'c',
    26: 'd',
    28: 'e',
    30: 'f',
    32: 'g',
    34: 'h',
}

entities = dict()


def gen_board(board_FEN):
    entities = {}
    start_board = board_FEN
    # 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR'
    x_coord = 20 #increment by 2
    y_coord = 30 #increment by 2

    for i in range(len(start_board)):
        if start_board[i] == '/':
            x_coord = 20
            y_coord += 2
            continue
        elif start_board[i].isdigit():
            for j in range(int(start_board[i])):
                x_coord += 2
            continue
        elif not start_board[i].isdigit():
            if start_board[i].isupper():
                piece_style = style.WHITE
                color_str = "White"
            else:
                piece_style = style.BLACK
                color_str = "Black"
            entities[color_str+rank[x_coord]+piece_name[start_board[i]]] = \
            Entity(x_coord, y_coord, pieces[start_board[i]], piece_style)
            #print(x_coord, y_coord, pieces[start_board[i]])
            x_coord += 2
            continue
        else:
            print("error parsing starting FEN")
            break

    #entities["WhatColor"] = Entity(40, 42, "You Play White", style.WHITE)

gen_board(chess.STARTING_BOARD_FEN)

#rng to decide color (for testing flip)
#playerColor = random.randint(1,2)

def game(state, press):
    mut_state = state
    if press == "q":
        mut_state["running"] = False



    return mut_state


run_loop(game, {"running": True, "entities": entities})
