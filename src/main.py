import sys, os, traceback, random, curses, chess, math, enum, itertools


## GLOBAL VARS ##

board = chess.Board()

#set to true to skip welcome screen
dev_mode = True

#Set true to disable post screen
post_screen_toggle = False

#prompt vars
prompt_x_coord = 1
prompt_y_coord = 1

#global strings
last_move_str = "no move yet"
user_input_string = ""
inputted_str = ""
status_str = ""
legal_move_str = ""
san_move_str = ""
history_arr = ["init"]
final_position = ""

move_amount = 0
game_outcome_enum = 0

#true if user hits enter key
entered_move = False
quit_from_welcome = False


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
    global dev_mode, post_screen_toggle
    key = 0
    cursor_x = 0
    cursor_y = 0
    stdscr = curses.initscr()
    height, width = stdscr.getmaxyx()

    # Clear and refresh the screen for a blank canvas
    stdscr.clear()
    stdscr.refresh()

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
    else:
        curses.init_pair(4, curses.COLOR_RED, curses.COLOR_WHITE)
        curses.init_pair(5, curses.COLOR_RED, curses.COLOR_BLACK)
        curses.init_pair(6, curses.COLOR_BLUE, curses.COLOR_WHITE)
        curses.init_pair(7, curses.COLOR_BLUE, curses.COLOR_BLACK)

    #move legality colors
    curses.init_pair(8, curses.COLOR_BLACK, curses.COLOR_GREEN)
    curses.init_pair(9, curses.COLOR_WHITE, curses.COLOR_RED)

    if not dev_mode:
        welcome_screen(stdscr)
        user_input_string = ""
    #start windows
    board_window = curses.newwin( math.floor((height/4)*3), math.floor(width/2), 0, 0)
    prompt_window = curses.newwin( math.floor((height)/4)-1 , math.floor(width/2),  math.floor((height/4)*3), 0)
    info_window = curses.newwin(math.floor(height/2), math.floor(width/2), 0, math.floor(width/2))
    history_window = curses.newwin( math.floor(height/2)-1, math.floor(width/2), math.floor(height/2), math.floor(width/2))

    windows_array = [board_window, info_window, prompt_window, history_window]

    # Loop where key is the last character pressed
    while (key != 15): # while not quitting (ctrl+o)
        if quit_from_welcome:
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
            prompt_window.resize(math.floor((height)/4)-1 , math.floor(width/2))
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
        cursor_x = max(0, cursor_x)
        cursor_x = min(width-1, cursor_x)

        cursor_y = max(0, cursor_y)
        cursor_y = min(height-1, cursor_y)

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
        stdscr.addstr(height-1, len(statusbarfull), " " * (width - len(statusbarfull) - 1))
        stdscr.attroff(curses.color_pair(3))

        for i in range(len(windows_array)):
            windows_array[i].border()

        ## EXTERNAL FUNCTION CALL !!! ###
        #external function calls
        update_input(prompt_window, key)
        board_input(board_window, key, width, height)
        game_logic(board_window)
        display_info(info_window)
        display_history(history_window)

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


#                        888          888                    d8b                            888
#                        888          888                    Y8P                            888
#                        888          888                                                   888
# 888  888 88888b.   .d88888  8888b.  888888 .d88b.          888 88888b.  88888b.  888  888 888888
# 888  888 888 "88b d88" 888     "88b 888   d8P  Y8b         888 888 "88b 888 "88b 888  888 888
# 888  888 888  888 888  888 .d888888 888   88888888         888 888  888 888  888 888  888 888
# Y88b 888 888 d88P Y88b 888 888  888 Y88b. Y8b.             888 888  888 888 d88P Y88b 888 Y88b.
#  "Y88888 88888P"   "Y88888 "Y888888  "Y888 "Y8888 88888888 888 888  888 88888P"   "Y88888  "Y888
#          888                                                            888
#          888                                                            888
#          888                                                            888

def update_input(prompt_window, key):
    global prompt_x_coord, prompt_y_coord, user_input_string, inputted_str, entered_move, status_str
    height, width = prompt_window.getmaxyx()

    if key==127: #delete key
        if prompt_x_coord-1 <= 0:
            delete_x = 1
        else:
            delete_x = prompt_x_coord-1
        prompt_window.addch(prompt_y_coord, delete_x, chr(8248)) #clear last char pointer
        prompt_window.addch(prompt_y_coord, delete_x+1, ' ') #clear last char printed
        prompt_x_coord -= 1 #decrement char position
        user_input_string = user_input_string[:-1]
    elif chr(key).isalnum() or key ==35 or key == 43:
        prompt_window.addch(prompt_y_coord, prompt_x_coord+1, chr(8248)) #indicate char youre on
        prompt_window.addch(prompt_y_coord, prompt_x_coord, key)
        prompt_x_coord += 1 #increment char position
        
    #adjust coordinates
    if prompt_x_coord <= 0:
        prompt_window.addch(prompt_y_coord, 1, ' ') #clear last char pointer
        prompt_x_coord = width-2
        prompt_y_coord -= 1
    if prompt_y_coord <= 0:
        prompt_x_coord = 1
        prompt_y_coord = 1        
    if prompt_x_coord >= width-1:
        prompt_x_coord = 1
        prompt_y_coord += 1
    if prompt_y_coord >= height-1:
        prompt_x_coord = width-2
        prompt_y_coord = height-2
        status_str = "char limit reached"
        return
        # for i in range(1, height-1):
        #     prompt_window.addstr(i, prompt_x_coord, " " * (width-1))
    if key==10: #enter key
        entered_move = True 
        inputted_str = user_input_string #set global string to check if move is legal
        user_input_string = "" #reset input buffer
        prompt_x_coord = 1 #reset char coordinates
        prompt_y_coord = 1#reset char coordinates
        #prompt_window.addch(prompt_y_coord, 0, '|')
        #prompt_window.addch(prompt_y_coord, 0, '>')

        for i in range(1, height-1): #clear window
            prompt_window.addstr(i, prompt_x_coord, " " * (width-1))
    
    #add to the current input buffer
    if key != 10 and key != 127 and (chr(key).isalnum() or key == 35 or key == 43): #not enter and not delete
        user_input_string += chr(key)

    #redraw border in case it was painted over
    prompt_window.border()
    prompt_window.addch(prompt_y_coord, 0, '>') #indicate line youre on

def board_input(board_window, key, screen_width, screen_height):
    global prompt_x_coord, prompt_y_coord, user_input_string, inputted_str, entered_move, status_str
    height, width = board_window.getmaxyx()

    if key != curses.KEY_MOUSE: #input needs to be mouse input
        return
    
    #(id_, x, y, z, bstate) = curses.getmouse()



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

def game_logic(board_window):
    global inputted_str, board, status_str, entered_move, last_move_str, history_arr, game_outcome_enum, move_amount
    inputted_str = inputted_str.strip(' ').strip('\0').strip('^@')
    legal_moves = generate_legal_moves()
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

                game_outcome_enum = game_outcome()
                if game_outcome_enum != 0:
                    status_str = outcome_tuple[game_outcome_enum]
                    final_position = board.fen
                    if not post_screen_toggle:
                        post_screen(draw_screen)
                    
                    

    #draw board
    draw_board(board_window, board.board_fen())
    legal_moves = generate_legal_moves()

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

def display_info(info_window):
    global last_move_str, status_str, inputted_str, legal_move_str, san_move_str
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
    san_move_str = "{}: {}".format("legal moves (san)", san_move_str)
    for y in range(5, height-1):
        wrap_y = y
        if len(san_move_str) > width-2:
            info_window.addstr(y, 1, san_move_str[:width-2])
            san_move_str = san_move_str[width-2:]
        else:
            info_window.addstr(y, 1, san_move_str)
            break

    legal_move_str = "{}: {}".format("legal moves (uci)", legal_move_str)
    for y in range(wrap_y+2, height-1):
        if len(legal_move_str) > width-2:
            info_window.addstr(y, 1, legal_move_str[:width-2])
            legal_move_str = legal_move_str[width-2:]
        else:
            info_window.addstr(y, 1, legal_move_str)
            break

    #info_window.addstr(7, 1, "{}: {}".format("legal moves (uci)", legal_move_str))
    info_window.attroff(curses.color_pair(8))

    status_str = ""



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
def display_history(history_window):
    global history_arr, move_amount
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

#      888                                       888                                    888
#      888                                       888                                    888
#      888                                       888                                    888
#  .d88888 888d888 8888b.  888  888  888         88888b.   .d88b.   8888b.  888d888 .d88888
# d88" 888 888P"      "88b 888  888  888         888 "88b d88""88b     "88b 888P"  d88" 888
# 888  888 888    .d888888 888  888  888         888  888 888  888 .d888888 888    888  888
# Y88b 888 888    888  888 Y88b 888 d88P         888 d88P Y88..88P 888  888 888    Y88b 888
#  "Y88888 888    "Y888888  "Y8888888P" 88888888 88888P"   "Y88P"  "Y888888 888     "Y88888
def draw_board(board_window, board_FEN):
    global board_square_coord
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
                board_window.attroff(curses.color_pair(color_pair))
                square_count += 1
                x_coord += x_inc
            continue

        elif not current_piece.isdigit():
            #determine proper color pair
            if current_piece.isupper():
                if square_count%2 == 0:
                    color_pair = 4
                else:
                    color_pair = 5
            else:
                if square_count%2 == 0:
                    color_pair = 6
                else:
                    color_pair = 7

            board_window.attron(curses.color_pair(color_pair))
            board_window.attron(curses.A_BOLD)

            board_window.addstr(y_coord, x_coord, pieces[current_piece.upper()]+" ")

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
def generate_legal_moves():
    global legal_move_str, san_move_str, board
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

def game_outcome():
    global game_outcome_enum

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


#                      888
#                       888
#                       888
#888  888  888  .d88b.  888  .d8888b .d88b.  88888b.d88b.   .d88b.         .d8888b   .d8888b 888d888 .d88b.   .d88b.  88888b.
#888  888  888 d8P  Y8b 888 d88P"   d88""88b 888 "888 "88b d8P  Y8b        88K      d88P"    888P"  d8P  Y8b d8P  Y8b 888 "88b
#888  888  888 88888888 888 888     888  888 888  888  888 88888888        "Y8888b. 888      888    88888888 88888888 888  888
#Y88b 888 d88P Y8b.     888 Y88b.   Y88..88P 888  888  888 Y8b.                 X88 Y88b.    888    Y8b.     Y8b.     888  888
# "Y8888888P"   "Y8888  888  "Y8888P "Y88P"  888  888  888  "Y8888 88888888 88888P'  "Y8888P 888     "Y8888   "Y8888  888  888



def welcome_screen(screen):
    global quit_from_welcome, user_input_string, inputted_str, entered_move
    height, width = screen.getmaxyx()
    key = 0

    prompt_welcome_window = curses.newwin( math.floor((height)/4)-1 , width,  math.floor((height/4)*3), 0)

    while (key != 12): # while not quitting
        if key == 15:
            quit_from_welcome = True
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

        update_input(prompt_welcome_window, key)

        prompt_welcome_window.border()
        screen.refresh()
        prompt_welcome_window.refresh()
        key = screen.getch()

    #reset global strings that may have been set in the prompt window
    user_input_string = ""
    inputted_str = ""
    entered_move = ""





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
def post_screen(screen1):
    global quit_from_post, user_input_string, inputted_str, entered_move
    screen1 = curses.initscr()
    height, width = screen1.getmaxyx()
    key = 0

    prompt_post_window = curses.newwin( math.floor((height)/4)-1 , width,  math.floor((height/4)*3), 0)
    board_post_window = curses.newwin( math.floor((height)-(height/3)), math.floor(width),  0, 0)
   
    while (key != 12): # while not quitting
        if key == 15:
            quit_from_post = True
            

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
        board_post_window.addstr(start_y + 3, (width // 2) - 2, '-' * 4)
        board_post_window.addstr(start_y + 5, start_x_final_history_str, final_history_str)
        draw_board(board_post_window, final_position)

        update_input(prompt_post_window, key)

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













if __name__ == "__main__":
    main()

# Centering calculations
# start_x_title = int((width // 2) - (len(title) // 2) - len(title) % 2)
# start_x_subtitle = int((width // 2) - (len(subtitle) // 2) - len(subtitle) % 2)
# start_x_keystr = int((width // 2) - (len(keystr) // 2) - len(keystr) % 2)
# start_y = int((height // 2) - 2)
