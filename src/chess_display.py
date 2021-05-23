import sys, os, traceback, random, curses, chess, math, enum, itertools

from chess_input import *
from game_logic import *



#GAME SCREEN WINDOWS 


#      888 d8b                   888                            d8b           .d888
#      888 Y8P                   888                            Y8P          d88P"
#      888                       888                                         888
#  .d88888 888 .d8888b  88888b.  888  8888b.  888  888          888 88888b.  888888 .d88b.
# d88" 888 888 88K      888 "88b 888     "88b 888  888          888 888 "88b 888   d88""88b
# 888  888 888 "Y8888b. 888  888 888 .d888888 888  888          888 888  888 888   888  888
# Y88b 888 888      X88 888 d88P 888 888  888 Y88b 888          888 888  888 888   Y88..88P
#  "Y88888 888  88888P' 88888P"  888 "Y888888  "Y88888 88888888 888 888  888 888    "Y88P"
#                       888                        888
#                       888                   Y8b d88P
#                       888                    "Y88P"
#display game information window for the game screen
def display_info(board, info_window, last_move_str, status_str, inputted_str, legal_move_str, san_move_str):
    #global last_move_str, status_str, inputted_str, legal_move_str, san_move_str
    height, width = info_window.getmaxyx()

    info_window.attron(curses.color_pair(3))
    if board.turn == chess.WHITE:
        info_window.addstr(1,1,"white to move")
    elif board.turn == chess.BLACK:
        #info_window.attron(curses.A_REVERSE)
        info_window.addstr(1,1,"black to move")
        #info_window.attroff(curses.A_REVERSE)

    info_window.addstr(2,1,"last move: {}".format(last_move_str))
    info_window.attroff(curses.color_pair(3))

    if status_str == "move is legal!":
        text_colour = 8
    else:
        text_colour = 9
    info_window.attron(curses.color_pair(text_colour))
    info_window.addstr(3, 1, status_str)
    info_window.attroff(curses.color_pair(text_colour))

    info_window.addstr(4, 1, "{}: {}".format("input",inputted_str))

    info_window.attron(curses.color_pair(8))

    #info_window.addstr(5, 1, "{}: {}".format("legal moves (san)", san_move_str))

    wrap_y = 0
    temp = san_move_str
    san_move_str = "{}: {}".format("legal moves (san)", san_move_str)
    for y in range(5, height-1):
        wrap_y = y
        if len(san_move_str) > width-2:
            info_window.addstr(y, 1, san_move_str[:width-2])
            san_move_str = san_move_str[width-2:]
        else:
            info_window.addstr(y, 1, san_move_str)
            break
    san_move_str = temp
    temp = legal_move_str
    legal_move_str = "{}: {}".format("legal moves (uci)", legal_move_str)
    for y in range(wrap_y+2, height-1):
        if len(legal_move_str) > width-2:
            info_window.addstr(y, 1, legal_move_str[:width-2])
            legal_move_str = legal_move_str[width-2:]
        else:
            info_window.addstr(y, 1, legal_move_str)
            break
    legal_move_str = temp
    #info_window.addstr(7, 1, "{}: {}".format("legal moves (uci)", legal_move_str))
    info_window.attroff(curses.color_pair(8))

    status_str = ""

    return (board, last_move_str, status_str, inputted_str, legal_move_str, san_move_str)



#          88  88                          88                                     88           88
#          88  ""                          88                                     88           ""               ,d
#          88                              88                                     88                            88
#  ,adPPYb,88  88  ,adPPYba,  8b,dPPYba,   88  ,adPPYYba,  8b       d8            88,dPPYba,   88  ,adPPYba,  MM88MMM  ,adPPYba,   8b,dPPYba,  8b       d8
# a8"    `Y88  88  I8[    ""  88P'    "8a  88  ""     `Y8  `8b     d8'            88P'    "8a  88  I8[    ""    88    a8"     "8a  88P'   "Y8  `8b     d8'
# 8b       88  88   `"Y8ba,   88       d8  88  ,adPPPPP88   `8b   d8'             88       88  88   `"Y8ba,     88    8b       d8  88           `8b   d8'
# "8a,   ,d88  88  aa    ]8I  88b,   ,a8"  88  88,    ,88    `8b,d8'              88       88  88  aa    ]8I    88,   "8a,   ,a8"  88            `8b,d8'
#  `"8bbdP"Y8  88  `"YbbdP"'  88`YbbdP"'   88  `"8bbdP"Y8      Y88'               88       88  88  `"YbbdP"'    "Y888  `"YbbdP"'   88              Y88'
#                             88                               d8'                                                                                 d8'
#                             88                              d8'     888888888888                                                                d8'
#display move history window for the game screen
def display_history(history_window, history_arr, move_amount, pieces):
    #global history_arr, move_amount
    height, width = history_window.getmaxyx()

    history_str_i = 0
    if len(history_arr) == 0:
        history_window.addstr(1, 1, "no moves yet")

    for y in range(1, height-1):
        if y >= len(history_arr):
            break
        hist_str = history_arr[history_str_i]
        piece_str = pieces["p"]
        if hist_str[0].isupper():
            piece_str = pieces[hist_str[0:1]]
            
        hist_str = "move "+str(move_amount-history_str_i)+": "+hist_str+" "+piece_str
        if len(hist_str) > width-2:
            history_window.addstr(y, 1, hist_str[:width-2])
            #hist_str = hist_str[width-2:]
        else:
            history_window.addstr(y, 1, hist_str)
        history_str_i += 1
    
    return (history_arr, move_amount)



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
            
            board_square_coord[key_tuple] = (floating_color, pieces[current_piece.upper()])

            board_window.attroff(curses.color_pair(color_pair))
            board_window.attroff(curses.A_BOLD)

            square_count += 1
            x_coord += x_inc
            continue
        else:
            print("error parsing starting FEN")
            break

    for i in range(8):
        board_window.addch(og_ycoord-y_inc-1, og_xcoord+x_inc*i, x_notation_string[i])
        board_window.addch(og_ycoord+8*y_inc+1, og_xcoord+x_inc*i, x_notation_string[i])
        board_window.addch(og_ycoord+y_inc*i, og_xcoord-x_inc-1, y_notation_string[i])
        board_window.addch(og_ycoord+y_inc*i, og_xcoord+8*x_inc+1, y_notation_string[i])
    
    return board_square_coord



#OTHER WINDOWS


#                      888
#                       888
#                       888
#888  888  888  .d88b.  888  .d8888b .d88b.  88888b.d88b.   .d88b.         .d8888b   .d8888b 888d888 .d88b.   .d88b.  88888b.
#888  888  888 d8P  Y8b 888 d88P"   d88""88b 888 "888 "88b d8P  Y8b        88K      d88P"    888P"  d8P  Y8b d8P  Y8b 888 "88b
#888  888  888 88888888 888 888     888  888 888  888  888 88888888        "Y8888b. 888      888    88888888 88888888 888  888
#Y88b 888 d88P Y8b.     888 Y88b.   Y88..88P 888  888  888 Y8b.                 X88 Y88b.    888    Y8b.     Y8b.     888  888
# "Y8888888P"   "Y8888  888  "Y8888P "Y88P"  888  888  888  "Y8888 88888888 88888P'  "Y8888P 888     "Y8888   "Y8888  888  888

#welcome screen that displays before the game screen 
def welcome_screen(screen, quit_game, user_input_string, inputted_str, entered_move, prompt_x_coord, prompt_y_coord, status_str):
    #global quit_game, user_input_string, inputted_str, entered_move
    height, width = screen.getmaxyx()
    key = 0

    prompt_welcome_window = curses.newwin( math.floor((height)/4)-1 , width,  math.floor((height/4)*3), 0)

    while (key != 12): # while not quitting
        if key == 15:
            quit_game = True
            break

        screen.clear()

        # Declaration of strings
        title = "chess-cli"[:width-1]
        subtitle = "play locally with a friend, against stockfish, or online with lichess!"[:width-1]
        keystr = "Last key pressed: {}".format(key)[:width-1]
        statusbarstr = "Press 'Ctrl-l' to skip to local | Press 'Ctrl-o' to quit"

        # Centering calculations
        start_x_title = int((width // 2) - (len(title) // 2) - len(title) % 2)
        start_x_subtitle = int((width // 2) - (len(subtitle) // 2) - len(subtitle) % 2)
        start_x_keystr = int((width // 2) - (len(keystr) // 2) - len(keystr) % 2)
        start_y = int((height // 2) - 2)

        # Rendering some text
        whstr = "Width: {}, Height: {}".format(width, height)
        screen.addstr(0, 0, whstr, curses.color_pair(1))

        # Render status bar
        screen.attron(curses.color_pair(3))
        screen.addstr(height-1, 0, statusbarstr)
        screen.addstr(height-1, len(statusbarstr), " " * (width - len(statusbarstr) - 1))
        screen.attroff(curses.color_pair(3))

        # Turning on attributes for title
        screen.attron(curses.color_pair(2))
        screen.attron(curses.A_BOLD)

        # Rendering title
        screen.addstr(start_y, start_x_title, title)

        # Turning off attributes for title
        screen.attroff(curses.color_pair(2))
        screen.attroff(curses.A_BOLD)

        # Print rest of text
        screen.addstr(start_y + 1, start_x_subtitle, subtitle)
        screen.addstr(start_y + 3, (width // 2) - 2, '-' * 4)
        screen.addstr(start_y + 5, start_x_keystr, keystr)

        update_input(prompt_welcome_window, key, prompt_x_coord, prompt_y_coord, user_input_string, inputted_str, entered_move, status_str)

        prompt_welcome_window.border()
        screen.refresh()
        prompt_welcome_window.refresh()
        key = screen.getch()

    #reset global strings that may have been set in the prompt window
    user_input_string = ""
    inputted_str = ""
    entered_move = ""

    return (quit_game, user_input_string, inputted_str, entered_move, prompt_x_coord, prompt_y_coord, status_str)





#########################################################################################################################
#                                 .                                                                           
#                                .o8                                                                           
# oo.ooooo.   .ooooo.   .oooo.o .o888oo              .oooo.o  .ooooo.  oooo d8b  .ooooo.   .ooooo.  ooo. .oo.   
#  888' `88b d88' `88b d88(  "8   888               d88(  "8 d88' `"Y8 `888""8P d88' `88b d88' `88b `888P"Y88b  
#  888   888 888   888 `"Y88b.    888               `"Y88b.  888        888     888ooo888 888ooo888  888   888  
#  888   888 888   888 o.  )88b   888 .             o.  )88b 888   .o8  888     888    .o 888    .o  888   888  
#  888bod8P' `Y8bod8P' 8""888P'   "888" ooooooooooo 8""888P' `Y8bod8P' d888b    `Y8bod8P' `Y8bod8P' o888o o888o 
#  888                                                                                                          
# o888o                                                                                                         
#########################################################################################################################                                                                    

#post game screen that displays after the game has reached a win condition
def post_screen(screen1, quit_game, user_input_string, inputted_str, entered_move, history_arr, final_position, prompt_x_coord, prompt_y_coord, status_str, board_square_coord, pieces):
    #global quit_game, user_input_string, inputted_str, entered_move, history_arr, final_position

    screen1.clear()
    screen1.refresh()

    height, width = screen1.getmaxyx()
    key = 0

    prompt_post_window = curses.newwin( math.floor((height)/4)-1 , width,  math.floor((height/4)*3), 0)
    board_post_window = curses.newwin( math.floor((height)-(height/3)), math.floor(width),  0, 0)
   
    while (key != 12): # while not quitting ctrl-l
        if key == 15: #ctrl-o
            quit_game = True
            break
        screen1.clear()

        # Declaration of strings
        title = "Game has ended."[:width-1]
        final_position_str = "Final position: "[:width-1]
        final_history_str = "Last key pressed: {}".format(key)[:width-1]
        statusbarstr = "Press 'Ctrl-l' to play again | Press 'Ctrl-o' to quit"

        # Centering calculations
        start_x_title = int((width // 2) - (len(title) // 2) - len(title) % 2)
        start_x_final_position_str = int((width // 2) - (len(final_position_str) // 2) - len(final_position_str) % 2)
        start_x_final_history_str = int((width // 2) - (len(final_history_str) // 2) - len(final_history_str) % 2)
        start_y = int((height // 2) - 2)

        # Render status bar
        screen1.attron(curses.color_pair(3))
        screen1.addstr(height-1, 0, statusbarstr)
        screen1.addstr(height-1, len(statusbarstr), " " * (width - len(statusbarstr) - 1))
        screen1.attroff(curses.color_pair(3))

        # Turning on attributes for title
        board_post_window.attron(curses.color_pair(2))
        board_post_window.attron(curses.A_BOLD)

        # Rendering title
        board_post_window.addstr(start_y, start_x_title, title)

        # Turning off attributes for title
        screen1.attroff(curses.color_pair(2))
        screen1.attroff(curses.A_BOLD)

        # Print rest of text
        board_post_window.addstr(start_y + 1, start_x_final_position_str, final_position_str)
        history = " -> ".join([str(elem) for elem in [ele for ele in reversed(history_arr)][1:]])[:width-2]
        board_post_window.addstr(start_y + 3, math.floor((width/2) - (len(history)/2)), history)

        board_post_window.addstr(start_y + 5, start_x_final_history_str, final_history_str)
        draw_board(board_post_window, final_position, board_square_coord, pieces)

        update_input(prompt_post_window, key, prompt_x_coord, prompt_y_coord, user_input_string, inputted_str, entered_move, status_str)

        prompt_post_window.border()
        board_post_window.border()
        screen1.refresh()
        prompt_post_window.refresh()
        board_post_window.refresh()
        key = screen1.getch()

    #reset global strings that may have been set in the prompt window
    user_input_string = ""
    inputted_str = ""
    entered_move = ""

    return (quit_game, user_input_string, inputted_str, entered_move, history_arr, final_position, prompt_x_coord, prompt_y_coord, status_str)


