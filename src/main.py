import sys, os, traceback, random, curses, chess, math, enum, itertools


from stockfish import Stockfish
from chess_input import *
from chess_display import *
from game_logic import *
from play_ai import *

## GLOBAL VARS ##
#set to true to skip welcome screen
skip_welcome = False

#Set true to disable post screen
post_screen_toggle = False

#quick game to get to post screen:
#f3 e5 g4 Qh4#

## FUNCTIONS ##

#                        d8b
#                        Y8P

# 88888b.d88b.   8888b.  888 88888b.
# 888 "888 "88b     "88b 888 888 "88b
# 888  888  888 .d888888 888 888  888
# 888  888  888 888  888 888 888  888
# 888  888  888 "Y888888 888 888  888

def main():
    curses.wrapper(draw_screen)

#      888
#      888
#      888
#  .d88888 888d888 8888b.  888  888  888        .d8888b   .d8888b 888d888 .d88b.   .d88b.  88888b.
# d88" 888 888P"      "88b 888  888  888        88K      d88P"    888P"  d8P  Y8b d8P  Y8b 888 "88b
# 888  888 888    .d888888 888  888  888        "Y8888b. 888      888    88888888 88888888 888  888
# Y88b 888 888    888  888 Y88b 888 d88P             X88 Y88b.    888    Y8b.     Y8b.     888  888
#  "Y88888 888    "Y888888  "Y8888888P" 88888888 88888P'  "Y8888P 888     "Y8888   "Y8888  888  888

def draw_screen(stdscr):
    global skip_welcome, post_screen_toggle

    board = chess.Board()
    #prompt vars
    # prompt_x_coord = 1
    # prompt_y_coord = 1


    move_amount = 0
    #@game_outcome_enum = 0
    key = 0

    #strings
    last_move_str = "no move yet"
    user_input_string = ""
    inputted_str = ""
    status_str = ""
    legal_move_str = ""
    san_move_str = ""
    history_arr = ["init"]
    final_position = ""
    floating_piece = ""

    #booleans
    entered_move = False #true if user hits enter key
    quit_game = False
    mouse_pressed = False
    floating = False
    ai_game = False

    outcome_tuple = (
        'Good luck.', #[0]
        'Checkmate!', #[1]
        'Stalemate.', #[2]
        'Draw - insufficient material.', #[3]
        'Draw - 75 move rule.', #[4]
        'Draw - fivefold repetition.', #[5]
        'Draw - 50 move rule.', #[6]
        'Draw by threefold repetition has been claimed.', #[7]
    )

    file = {
        'a': 0,
        'b': 1,
        'c': 2,
        'd': 3,
        'e': 4,
        'f': 5,
        'g': 6,
        'h': 7,
    }

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
        #'p': '♟︎',
        'p': '♙',

    }

    board_square_coord = dict()

    
    #cursor_x = 0
    #cursor_y = 0
    stdscr = curses.initscr()
    height, width = stdscr.getmaxyx()

    # Clear and refresh the screen for a blank canvas
    stdscr.clear()
    stdscr.refresh()

    #necessary for mouse input, start keypad, read all mouse events
    #stdscr.keypad(1)
    #curses.mousemask(curses.ALL_MOUSE_EVENTS | curses.REPORT_MOUSE_POSITION)
    #print("\033[?1003h")

    # allow input, Start colors in curses
    curses.echo()
    curses.curs_set(0);
    curses.start_color()

    #curses.use_default_colors()
    curses.init_pair(1, curses.COLOR_CYAN, curses.COLOR_BLACK)
    curses.init_pair(2, curses.COLOR_RED, curses.COLOR_BLACK)
    curses.init_pair(3, curses.COLOR_BLACK, curses.COLOR_WHITE)

    #piece and square colors
    if curses.can_change_color():
        light_square = 215 #SandyBrown
        dark_square = 94 #Orange4
        light_piece = 230 #Cornsilk1
        dark_piece = 233 #Grey7

        curses.init_pair(4, light_piece, light_square)
        curses.init_pair(5, light_piece, dark_square)
        curses.init_pair(6, dark_piece, light_square)
        curses.init_pair(7, dark_piece, dark_square)

        #floating piece colors
        curses.init_pair(10, light_piece, dark_piece)
        curses.init_pair(11, dark_piece, light_piece)
    else:
        curses.init_pair(4, curses.COLOR_RED, curses.COLOR_WHITE)
        curses.init_pair(5, curses.COLOR_RED, curses.COLOR_BLACK)
        curses.init_pair(6, curses.COLOR_BLUE, curses.COLOR_WHITE)
        curses.init_pair(7, curses.COLOR_BLUE, curses.COLOR_BLACK)

    #move legality colors
    curses.init_pair(8, curses.COLOR_BLACK, curses.COLOR_GREEN)
    curses.init_pair(9, curses.COLOR_WHITE, curses.COLOR_RED)



    if not skip_welcome:
        #welcome screen
        quit_game, user_input_string, inputted_str, entered_move, status_str, ai_game = \
            welcome_screen( stdscr, quit_game, user_input_string, inputted_str, entered_move, status_str, ai_game ) 
            
    if ai_game:
        stockfish_obj = Stockfish(parameters={"Threads": 2, "Minimum Thinking Time": 30})
    #start windows
    board_window = curses.newwin( math.floor((height/4)*3), \
                   math.floor(width/2), 0, 0)
    prompt_window = curses.newwin( math.floor((height)/4)-1 , \
                    math.floor(width/2),  math.floor((height/4)*3), 0)
    info_window = curses.newwin(math.floor(height/2), math.floor(width/2), 0, \
                  math.floor(width/2))
    history_window = curses.newwin( math.floor(height/2)-1, \
                     math.floor(width/2), math.floor(height/2), math.floor(width/2))

    windows_array = [board_window, info_window, prompt_window, history_window]

    # Loop where key is the last character pressed
    while (key != 15): # while not quitting (ctrl+o)
        if quit_game:
            break
        # Initialization
        stdscr.clear()
        board_window.clear()
        info_window.clear()
        history_window.clear()

        #resize everything if necessary
        if curses.is_term_resized(height, width):
            height, width = stdscr.getmaxyx() #get new height and width

            #resize the terminal and refresh
            curses.resize_term(height, width)
            stdscr.clear()
            stdscr.refresh()

            #resize windows based on new dimensions
            board_window.resize(math.floor((height/4)*3), math.floor(width/2))
            prompt_window.resize(math.floor((height)/4)-1 , \
            math.floor(width/2))
            info_window.resize(math.floor(height/2), math.floor(width/2))
            history_window.resize(math.floor(height/2)-1, math.floor(width/2))

            #move windows to appropriate locations
            board_window.mvwin(0, 0)
            prompt_window.mvwin(math.floor((height/4)*3), 0)
            info_window.mvwin(0, math.floor(width/2))
            history_window.mvwin(math.floor(height/2), math.floor(width/2))

            #clear and refresh all windows
            for i in range(len(windows_array)):
                windows_array[i].clear()
                windows_array[i].refresh()

        #get window dimensions
        height, width = stdscr.getmaxyx()
        board_window_height, board_window_width = board_window.getmaxyx()
        info_window_height, info_window_width = info_window.getmaxyx()
        prompt_window_height, prompt_window_width = prompt_window.getmaxyx()
        history_window_height, history_window_width = history_window.getmaxyx()

        #get mouse location
        #cursor_x = max(0, cursor_x)
        #cursor_x = min(width-1, cursor_x)

        #cursor_y = max(0, cursor_y)
        #cursor_y = min(height-1, cursor_y)

        # Declaration of strings
        board_title = "board"[:width-1]
        info_title = "info"[:width-1]
        prompt_title = "prompt"[:width-1]
        history_title = "move_history"[:width-1]

        keystr = "Last key pressed: {}".format(key)[:width-1]
        #statusbarstr = "Press 'Ctrl+o' to exit | CHESS-CLI | Pos: {}, {}".format(cursor_x, cursor_y)
        statusbarstr = "Press 'Ctrl+o' to exit | CHESS-CLI"

        statusbarfull = "{} | {}".format(statusbarstr, keystr)
        #statusbarfull = ""

        if key == 0:
            keystr = "No key press detected..."[:width-1]

        # Render status bar
        stdscr.attron(curses.color_pair(3))
        stdscr.addstr(height-1, 0, statusbarfull)
        stdscr.addstr(height-1, len(statusbarfull), " " * \
        (width - len(statusbarfull) - 1))
        stdscr.attroff(curses.color_pair(3))

        for i in range(len(windows_array)):
            windows_array[i].border()

        ## EXTERNAL FUNCTION CALL !!! ###
        #external function calls

        #update_input updates the game screen prompt window and returns what the user is currently typing
        user_input_string, inputted_str, entered_move, status_str = \
            update_input(prompt_window, key, user_input_string, inputted_str, \
        entered_move, status_str)
         #update the board window mouse input
        #mouse_pressed, floating_piece, floating = board_window_mouse_input(board_window, key, width, height, board_square_coord, mouse_pressed, floating_piece, floating)
       
        #game_logic determines if an inputted move is legal and manages the gamestate

        if ai_game:
            #call play_stockfish
            inputted_str, board, status_str, entered_move, last_move_str, \
            history_arr, move_amount, final_position,\
            post_screen_toggle, board_square_coord, legal_move_str, san_move_str, stockfish_obj = \
                stockfish_logic( board_window, inputted_str, board, 
            status_str, entered_move, last_move_str, history_arr, \
            move_amount, final_position, \
            post_screen_toggle, board_square_coord, pieces, \
            legal_move_str, san_move_str, outcome_tuple, stockfish_obj)

        else:
            #call local_game_logic
            inputted_str, board, status_str, entered_move, last_move_str, \
            history_arr, move_amount, final_position, \
            post_screen_toggle, board_square_coord, legal_move_str, san_move_str = \
                local_game_logic(board_window, inputted_str, board, \
            status_str, entered_move, last_move_str, history_arr, \
             move_amount, final_position, \
            post_screen_toggle, board_square_coord, pieces, \
            legal_move_str, san_move_str, outcome_tuple)
        
        if post_screen_toggle: #check if post_screen is enabled
            post_screen_toggle = False

            #post_screen displays after the win condition has been met
            quit_game, user_input_string, inputted_str, entered_move, \
            history_arr, final_position, \
            status_str = post_screen(stdscr, quit_game, user_input_string, \
            inputted_str, entered_move, history_arr, final_position, \
            status_str, board_square_coord, \
            pieces)
            if quit_game:
                break

            #return to the welcome screen
            quit_game, user_input_string, inputted_str, entered_move, \
             status_str = \
            welcome_screen(stdscr, quit_game, user_input_string, inputted_str, \
            entered_move, status_str)
            continue
        
        #windows for the game screen

        #display game information
        status_str, legal_move_str, san_move_str = \
            display_info(board, info_window, last_move_str, status_str, inputted_str, legal_move_str, san_move_str)
        #display move history
        display_history(history_window, history_arr, move_amount, pieces)

        #end of external function call section 

        # Turning on attributes for title
        for i in range(len(windows_array)):
            windows_array[i].attron(curses.color_pair(2))
            windows_array[i].attron(curses.A_BOLD)

        # Rendering title
        board_window.addstr(0, 1, board_title)
        info_window.addstr(0, 1, info_title)
        prompt_window.addstr(0, 1, prompt_title)
        history_window.addstr(0,1, history_title)

        # Turning off attributes for title
        for i in range(len(windows_array)):
            windows_array[i].attroff(curses.color_pair(2))
            windows_array[i].attroff(curses.A_BOLD)

        # Refresh the screen
        stdscr.refresh()
        for i in range(len(windows_array)):
            windows_array[i].refresh()

        # Wait for next input
        key = stdscr.getch()




if __name__ == "__main__":
    main()

# Centering calculations
# start_x_title = int((width // 2) - (len(title) // 2) - len(title) % 2)
# start_x_subtitle = int((width // 2) - (len(subtitle) // 2) - len(subtitle) % 2)
# start_x_keystr = int((width // 2) - (len(keystr) // 2) - len(keystr) % 2)
# start_y = int((height // 2) - 2)
