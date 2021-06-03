import sys, os, traceback, random, curses, chess, math, enum, itertools, stockfish

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
def local_game_logic( board_window, inputted_str, board, status_str, entered_move, \
                last_move_str, history_arr, \
                move_amount, final_position, post_screen_toggle, 
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

                game_outcome_enum = game_outcome(board)
                if game_outcome_enum != 0:
                    status_str = outcome_tuple[game_outcome_enum]
                    final_position = board.board_fen()
                    post_screen_toggle = True

    #draw board
    board_square_coord = draw_board(board_window, board.board_fen(), \
                         board_square_coord, pieces)
    legal_moves = generate_legal_moves(legal_move_str, san_move_str, board)


    return (inputted_str, board, status_str, entered_move, last_move_str, \
            history_arr, move_amount, final_position,\
            post_screen_toggle, board_square_coord, legal_move_str, \
            san_move_str)




#      888                                       888                                    888
#      888                                       888                                    888
#      888                                       888                                    888
#  .d88888 888d888 8888b.  888  888  888         88888b.   .d88b.   8888b.  888d888 .d88888
# d88" 888 888P"      "88b 888  888  888         888 "88b d88""88b     "88b 888P"  d88" 888
# 888  888 888    .d888888 888  888  888         888  888 888  888 .d888888 888    888  888
# Y88b 888 888    888  888 Y88b 888 d88P         888 d88P Y88..88P 888  888 888    Y88b 888
#  "Y88888 888    "Y888888  "Y8888888P" 88888888 88888P"   "Y88P"  "Y888888 888     "Y88888

# function called from game_logic, draws the board for board window on the game screen
def draw_board(board_window, board_FEN, board_square_coord, pieces):
    #global board_square_coord
    height, width = board_window.getmaxyx()
    board_square_coord = {}
    x_notation_string = 'abcdefgh'
    y_notation_string = '87654321'
    # 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR'
    x_inc = 2
    y_inc = 1

    x_coord = width//2 - 4*x_inc #increment by 2
    y_coord = height//2 - 4*y_inc #increment by 2

    og_xcoord = x_coord
    og_ycoord = y_coord

    square_count = 0

    for i in range(len(board_FEN)): #loop to parse the FEN stirng

        current_piece = board_FEN[i] #current piece character from the FEN string
        key_tuple = (x_coord, y_coord)

        if current_piece == '/':
            board_window.addstr(y_coord, x_coord, chr(9))
            x_coord = og_xcoord #set x_coord to first in the row
            y_coord += y_inc #incremen
            square_count += 1
            continue

        elif current_piece.isdigit():
            for j in range(int(current_piece)):
                if square_count%2 == 0:
                    color_pair = 4
                else:
                    color_pair = 5
                board_window.attron(curses.color_pair(color_pair))
                board_window.addstr(y_coord, x_coord, " "+chr(9)) #add a space+tab character for an empty square
                board_square_coord[key_tuple] = (color_pair, None)
                board_window.attroff(curses.color_pair(color_pair))
                square_count += 1
                x_coord += x_inc
            continue

        elif not current_piece.isdigit():
            #determine proper color pair
            if current_piece.isupper():
                floating_color = 10
                if square_count%2 == 0:
                    color_pair = 4
                else:
                    color_pair = 5
            else:
                floating_color = 11
                if square_count%2 == 0:
                    color_pair = 6
                else:
                    color_pair = 7

            board_window.attron(curses.color_pair(color_pair))
            board_window.attron(curses.A_BOLD)

            board_window.addstr(y_coord, x_coord, pieces[current_piece.upper()]+" ")
            
            board_square_coord[key_tuple] = \
                (floating_color, pieces[current_piece.upper()])

            board_window.attroff(curses.color_pair(color_pair))
            board_window.attroff(curses.A_BOLD)

            square_count += 1
            x_coord += x_inc
            continue
        else:
            print("error parsing starting FEN")
            break

    for i in range(8):
        board_window.addch(og_ycoord-y_inc-1, \
        og_xcoord+x_inc*i, x_notation_string[i])
        board_window.addch(og_ycoord+8*y_inc+1, \
        og_xcoord+x_inc*i, x_notation_string[i])
        board_window.addch(og_ycoord+y_inc*i, \
        og_xcoord-x_inc-1, y_notation_string[i])
        board_window.addch(og_ycoord+y_inc*i, \
        og_xcoord+8*x_inc+1, y_notation_string[i])
    
    return board_square_coord


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
def game_outcome(board):
    game_outcome_enum = 0
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
