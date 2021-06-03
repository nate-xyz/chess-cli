import sys, os, traceback, random, curses, chess, math, enum, itertools, stockfish

from game_logic import *
                                                                                                                                                                                                                      
#                                             88           ad88  88             88                                                                                        88                            88              
#              ,d                             88          d8"    ""             88                                                                                        88                            ""              
#              88                             88          88                    88                                                                                        88                                            
# ,adPPYba,  MM88MMM  ,adPPYba,    ,adPPYba,  88   ,d8  MM88MMM  88  ,adPPYba,  88,dPPYba,              ,adPPYb,d8  ,adPPYYba,  88,dPYba,,adPYba,    ,adPPYba,            88   ,adPPYba,    ,adPPYb,d8  88   ,adPPYba,  
# I8[    ""    88    a8"     "8a  a8"     ""  88 ,a8"     88     88  I8[    ""  88P'    "8a            a8"    `Y88  ""     `Y8  88P'   "88"    "8a  a8P_____88            88  a8"     "8a  a8"    `Y88  88  a8"     ""  
#  `"Y8ba,     88    8b       d8  8b          8888[       88     88   `"Y8ba,   88       88            8b       88  ,adPPPPP88  88      88      88  8PP"""""""            88  8b       d8  8b       88  88  8b          
# aa    ]8I    88,   "8a,   ,a8"  "8a,   ,aa  88`"Yba,    88     88  aa    ]8I  88       88            "8a,   ,d88  88,    ,88  88      88      88  "8b,   ,aa            88  "8a,   ,a8"  "8a,   ,d88  88  "8a,   ,aa  
# `"YbbdP"'    "Y888  `"YbbdP"'    `"Ybbd8"'  88   `Y8a   88     88  `"YbbdP"'  88       88             `"YbbdP"Y8  `"8bbdP"Y8  88      88      88   `"Ybbd8"'            88   `"YbbdP"'    `"YbbdP"Y8  88   `"Ybbd8"'  
#                                                                                                       aa,    ,88                                                                          aa,    ,88                  
#                                                                                          888888888888  "Y8bbdP"                                             888888888888                   "Y8bbdP"                          


def stockfish_logic( board_window, inputted_str, board, status_str, \
                    entered_move, last_move_str, history_arr, \
                    game_outcome_enum, move_amount, final_position, \
                    post_screen_toggle, board_square_coord, pieces, \
                    legal_move_str, san_move_str, outcome_tuple, stockfish_obj):
    #global inputted_str, board, status_str, entered_move, last_move_str, history_arr, game_outcome_enum, move_amount, final_position, post_screen_toggle
    inputted_str = inputted_str.strip(' ').strip('\0').strip('^@')
    legal_moves = generate_legal_moves(legal_move_str, san_move_str, board)
    legal_moves_san = legal_moves[0] 
    legal_moves_san_lowercase = legal_moves[1]
    legal_moves_uci = legal_moves[2]

    legal_moves = list(itertools.chain.from_iterable(legal_moves))

    stockfish_obj.set_elo_rating(1350) #implement user choice later
    key = 0

    player_color = random.randint (1, 2)
    if player_color == 1:
        player_control = chess.WHITE
    else:
        player_control = not chess.WHITE
    while (key != 15): #while not quitting
        if player_control == True:
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

                 
        else:
            inputted_str = stockfish_obj.get_best_move_time(1000) 
            #take 1 second to think, set to inputted_str. stockfish will format like a2a3, e2e4 etc (uci)
            if inputted_str not in legal_moves:
                status_str = "last input is invalid"
            else:
                status_str = "move is legal!"
                if board.is_legal(inputted_str):
                    board.push_uci(inputted_str) #make the actual move with the chess module
                    last_move_str = (board.parse_san(inputted_str)) #set stockfish move to san for history
                    history_arr.insert(0, last_move_str) #push to the front of the history stack for the history window
                    move_amount+=1 #increment the global move amount for the history window
                    curses.flash()
                    curses.beep()


        game_outcome_enum = game_outcome(board, game_outcome_enum)
        if game_outcome_enum != 0:
            status_str = outcome_tuple[game_outcome_enum]
            final_position = board.board_fen()
            post_screen_toggle = True

    #draw board
    board_square_coord = draw_board(board_window, board.board_fen(), \
                         board_square_coord, pieces)
    legal_moves = generate_legal_moves(legal_move_str, san_move_str, board)


    return (inputted_str, board, status_str, entered_move, last_move_str, \
            history_arr, game_outcome_enum, move_amount, final_position,\
            post_screen_toggle, board_square_coord, legal_move_str, \
            san_move_str, stockfish_obj)




