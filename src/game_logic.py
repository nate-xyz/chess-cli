import sys, os, traceback, random, curses, chess, math, enum, itertools

from chess_display import draw_board

#                                                  888                   d8b
#                                                  888                   Y8P
#                                                  888
#  .d88b.   8888b.  88888b.d88b.   .d88b.          888  .d88b.   .d88b.  888  .d8888b
# d88P"88b     "88b 888 "888 "88b d8P  Y8b         888 d88""88b d88P"88b 888 d88P"
# 888  888 .d888888 888  888  888 88888888         888 888  888 888  888 888 888
# Y88b 888 888  888 888  888  888 Y8b.             888 Y88..88P Y88b 888 888 Y88b.
#  "Y88888 "Y888888 888  888  888  "Y8888 88888888 888  "Y88P"   "Y88888 888  "Y8888P
#      888                                                           888
# Y8b d88P                                                      Y8b d88P
#  "Y88P"                                                        "Y88P"

#game_logic determines if an inputted move is legal and manages the gamestate
def game_logic( board_window, inputted_str, board, status_str, entered_move, last_move_str, history_arr, game_outcome_enum, move_amount, final_position, post_screen_toggle, 
                board_square_coord, pieces, 
                legal_move_str, san_move_str,
                outcome_tuple):
    #global inputted_str, board, status_str, entered_move, last_move_str, history_arr, game_outcome_enum, move_amount, final_position, post_screen_toggle
    inputted_str = inputted_str.strip(' ').strip('\0').strip('^@')
    legal_moves = generate_legal_moves(legal_move_str, san_move_str, board)
    legal_moves_san = legal_moves[0] 
    legal_moves_san_lowercase = legal_moves[1]
    legal_moves_uci = legal_moves[2]

    legal_moves = list(itertools.chain.from_iterable(legal_moves))


    if entered_move:
        entered_move = False
        if inputted_str == 'undo':
            board.pop()
        else:
            if inputted_str not in legal_moves:
                status_str = "last input is invalid"
            else:
                status_str = "move is legal!"

                if inputted_str in legal_moves_san_lowercase:
                    inputted_str = legal_moves_san[legal_moves_san_lowercase.index(inputted_str)] #get the equivalent string with proper case
  
                if board.is_legal(board.parse_san(inputted_str)):
                    board.push_san(inputted_str) #make the actual move with the chess module
                    last_move_str = inputted_str #set the last move string to be displayed in the info window
                    history_arr.insert(0, inputted_str) #push to the front of the history stack for the history window
                    move_amount+=1 #increment the global move amount for the history window
                    curses.flash()
                    curses.beep()

                game_outcome_enum = game_outcome(board, game_outcome_enum)
                if game_outcome_enum != 0:
                    status_str = outcome_tuple[game_outcome_enum]
                    final_position = board.board_fen()
                    post_screen_toggle = True

    #draw board
    board_square_coord = draw_board(board_window, board.board_fen(), board_square_coord, pieces)
    legal_moves = generate_legal_moves(legal_move_str, san_move_str, board)


    return (inputted_str, board, status_str, entered_move, last_move_str, history_arr, game_outcome_enum, move_amount, final_position, post_screen_toggle, board_square_coord, legal_move_str, san_move_str)

#                                                     888                    888                            888
#                                                     888                    888                            888
#                                                     888                    888                            888
#  .d88b.   .d88b.  88888b.   .d88b.  888d888 8888b.  888888 .d88b.          888  .d88b.   .d88b.   8888b.  888          88888b.d88b.   .d88b.  888  888  .d88b.  .d8888b
# d88P"88b d8P  Y8b 888 "88b d8P  Y8b 888P"      "88b 888   d8P  Y8b         888 d8P  Y8b d88P"88b     "88b 888          888 "888 "88b d88""88b 888  888 d8P  Y8b 88K
# 888  888 88888888 888  888 88888888 888    .d888888 888   88888888         888 88888888 888  888 .d888888 888          888  888  888 888  888 Y88  88P 88888888 "Y8888b.
# Y88b 888 Y8b.     888  888 Y8b.     888    888  888 Y88b. Y8b.             888 Y8b.     Y88b 888 888  888 888          888  888  888 Y88..88P  Y8bd8P  Y8b.          X88
#  "Y88888  "Y8888  888  888  "Y8888  888    "Y888888  "Y888 "Y8888 88888888 888  "Y8888   "Y88888 "Y888888 888 88888888 888  888  888  "Y88P"    Y88P    "Y8888   88888P'
#      888                                                                                     888
# Y8b d88P                                                                                Y8b d88P
#  "Y88P"                                                                                  "Y88P"

#helper function for game logic, generates 3 arrays of legal moves based on the current gamestate. 
#the first array is the current legal SAN moves, the second is lowercase SAN, and the third is UCI moves
def generate_legal_moves(legal_move_str, san_move_str, board):
    #global legal_move_str, san_move_str, board
    legal_moves = [[],[],[]]
    legal_moves_san = []
    legal_moves_san_lowercase = []
    legal_moves_uci = []
    legal_move_str = ""
    san_move_str = ""
    for move in board.legal_moves:

        #append to legal moves array
        legal_moves_san.append(board.san(move))
        legal_moves_san_lowercase.append(board.san(move).lower())
        legal_moves_uci.append(chess.Move.uci(move))

        #add to the global strings to be displayed in the info window
        legal_move_str += chess.Move.uci(move) + " "
        san_move_str += board.san(move) + " "

        # piece_char = board.piece_at( file[movo_str[0]] + (int(movo_str[1])- 1)\
        #  * 8 ).symbol()
        # if piece_char.upper() == "P":
        #     piece_char = ""
        # san_move_str += piece_char + movo_str[2:4] + " "

    #return legal moves array
    legal_moves[0] = legal_moves_san
    legal_moves[1] = legal_moves_san_lowercase
    legal_moves[2] = legal_moves_uci
    return legal_moves

#          █████                                                              ████
#         ░░███                                                              ░░███
 # ██████  ░███████    ██████   █████   █████            ████████  █████ ████ ░███   ██████   █████
 #███░░███ ░███░░███  ███░░███ ███░░   ███░░            ░░███░░███░░███ ░███  ░███  ███░░███ ███░░
#░███ ░░░  ░███ ░███ ░███████ ░░█████ ░░█████            ░███ ░░░  ░███ ░███  ░███ ░███████ ░░█████
#░███  ███ ░███ ░███ ░███░░░   ░░░░███ ░░░░███           ░███      ░███ ░███  ░███ ░███░░░   ░░░░███
#░░██████  ████ █████░░██████  ██████  ██████  █████████ █████     ░░████████ █████░░██████  ██████
# ░░░░░░  ░░░░ ░░░░░  ░░░░░░  ░░░░░░  ░░░░░░  ░░░░░░░░░ ░░░░░       ░░░░░░░░ ░░░░░  ░░░░░░  ░░░░░░

#returns an enumerated type based on the current game outcome
def game_outcome(board, game_outcome_enum):
    #global game_outcome_enum

    if board.is_checkmate():
        game_outcome_enum = 1
    if board.is_stalemate():
        game_outcome_enum = 2
    if board.is_insufficient_material():
        game_outcome_enum = 3
    if board.is_seventyfive_moves():
        game_outcome_enum = 4
    if board.is_fivefold_repetition():
        game_outcome_enum = 5

    return game_outcome_enum
