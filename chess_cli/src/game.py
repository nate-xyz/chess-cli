import chess

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

class Game:

    def __init__(self, initial_placement):
        self.game_board = Board(self, initial_placement)

class Board:


    def __init__(self, parent_instance, initial_placement):
        """Constructor for board class
        """

        self.parent_instance = parent_instance
        initial_x1 = initial_placement[0][0]
        initial_y1 = initial_placement[0][1]
        initial_val1 = initial_placement[0][2]
        initial_x2 = initial_placement[1][0]
        initial_y2 = initial_placement[1][1]
        initial_val2 = initial_placement[1][2]
        self.board_positions = [[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]]
        self.board_positions[initial_y1][initial_x1] = initial_val1
        self.board_positions[initial_y2][initial_x2] = initial_val2

    def print_board(self):
        """Debug function for printing board
        """

        for i in range(0, len(self.board_positions)):
            print(self.board_positions[i])

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
                    piece_style = self.set_color(py_cui.BLACK_ON_WHITE)
                else:
                    piece_style = self.set_color(py_cui.BLACK_ON_CYAN)

                entities[color_str+rank[x_coord]+piece_name[start_board[i]]] = \
                Entity(x_coord, y_coord, pieces[start_board[i]], piece_style)
                #print(x_coord, y_coord, pieces[start_board[i]])
                x_coord += 2
                continue
            else:
                print("error parsing starting FEN")
                break



def gen_board(board_FEN):
    #entities = {}
    start_board = board_FEN
    # 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR'
    x_coord = 20 #increment by 2
    y_coord = 30 #increment by 2



    #entities["WhatColor"] = Entity(40, 42, "You Play White", style.WHITE)

gen_board(chess.STARTING_BOARD_FEN)
















    
