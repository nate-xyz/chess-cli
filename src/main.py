import sys, os, traceback, random, curses, chess, math

board = chess.Board()

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

#true if user hits enter key
entered_move = False

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

    #start windows
    board_window = curses.newwin( math.floor((height/4)*3), math.floor(width/2), 0, 0)
    prompt_window = curses.newwin( math.floor((height)/4)-1 , math.floor(width/2),  math.floor((height/4)*3), 0)
    info_window = curses.newwin(math.floor(height/2), math.floor(width/2), 0, math.floor(width/2))
    history_window = curses.newwin( math.floor(height/2)-1, math.floor(width/2), math.floor(height/2), math.floor(width/2))

    windows_array = [board_window, info_window, prompt_window, history_window]
    # Loop where key is the last character pressed
    while (key != 15): # while not quitting (ctrl+o)

        # Initialization
        stdscr.clear()
        board_window.clear()
        info_window.clear()

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

        # Turning on attributes for title
        for i in range(len(windows_array)):
            windows_array[i].border()
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

        #external function calls
        update_input(prompt_window, key)
        game_logic(board_window)
        display_info(info_window)

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
    global prompt_x_coord, prompt_y_coord, user_input_string, inputted_str, entered_move
    height, width = prompt_window.getmaxyx()

    prompt_window.addch(prompt_y_coord, prompt_x_coord, key)
    prompt_window.addch(prompt_y_coord, 0, '>')
    prompt_window.addch(prompt_y_coord, prompt_x_coord+1, chr(8248))

    prompt_x_coord += 1

    if prompt_x_coord >= width-1:
        prompt_x_coord = 1
        prompt_y_coord += 1

    if prompt_y_coord >= height-1:
        prompt_x_coord = 1
        prompt_y_coord = 1
        for i in range(1, height-1):
            prompt_window.addstr(i, prompt_x_coord, " " * (width-1))

    if key==10: #enter key
        entered_move = True
        inputted_str = user_input_string
        user_input_string = ""
        prompt_window.addch(prompt_y_coord, 0, '|')
        prompt_x_coord = 1
        prompt_y_coord = 1
        prompt_window.addch(prompt_y_coord, 0, '>')

        for i in range(1, height-1):
            prompt_window.addstr(i, prompt_x_coord, " " * (width-1))
    else:
        user_input_string += chr(key)

#     prompt_x_coord = 0
#     prompt_y_coord = 0


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
    global inputted_str, board, status_str, entered_move, last_move_str
    inputted_str = inputted_str.strip(' ').strip('\0').strip('^@')
    legal_moves = []
    legal_moves = generate_legal_moves()
    if entered_move:
        entered_move = False
        if inputted_str == 'undo':
            board.pop()
        else:
            if inputted_str not in legal_moves:
                status_str = "last input is invalid"
            else:
                status_str = "move is legal!"
                if board.is_legal(board.parse_san(inputted_str)):
                    last_move_str = (board.parse_san(inputted_str))
                    curses.flash()
                    curses.beep()
                    board.push_san(inputted_str)
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

    san_move_str = "{}: {}".format("legal moves (san)", san_move_str)
    for y in range(5, height-1):
        if len(san_move_str) > width-2:
            info_window.addstr(y, 1, san_move_str[:width-2])
            san_move_str = san_move_str[width-2:]
        else:
            info_window.addstr(y, 1, san_move_str)
            break
    #info_window.addstr(7, 1, "{}: {}".format("legal moves (uci)", legal_move_str))
    info_window.attroff(curses.color_pair(8))

    status_str = ""

#      888                                       888                                    888
#      888                                       888                                    888
#      888                                       888                                    888
#  .d88888 888d888 8888b.  888  888  888         88888b.   .d88b.   8888b.  888d888 .d88888
# d88" 888 888P"      "88b 888  888  888         888 "88b d88""88b     "88b 888P"  d88" 888
# 888  888 888    .d888888 888  888  888         888  888 888  888 .d888888 888    888  888
# Y88b 888 888    888  888 Y88b 888 d88P         888 d88P Y88..88P 888  888 888    Y88b 888
#  "Y88888 888    "Y888888  "Y8888888P" 88888888 88888P"   "Y88P"  "Y888888 888     "Y88888
def draw_board(board_window, board_FEN):
    height, width = board_window.getmaxyx()

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

    for i in range(len(board_FEN)):
        current_piece = board_FEN[i]

        if current_piece == '/':
            board_window.addstr(y_coord, x_coord, chr(9))
            x_coord = og_xcoord
            y_coord += y_inc
            square_count += 1
            continue
        elif current_piece.isdigit():

            for j in range(int(current_piece)):
                if square_count%2 == 0:
                    color_pair = 4
                else:
                    color_pair = 5
                board_window.attron(curses.color_pair(color_pair))
                board_window.addstr(y_coord, x_coord, " "+chr(9))
                board_window.attroff(curses.color_pair(color_pair))
                square_count += 1
                x_coord += x_inc
            continue
        elif not current_piece.isdigit():


            #determine proper color pair

            # curses.init_pair(4, curses.COLOR_RED, curses.COLOR_WHITE)
            # curses.init_pair(5, curses.COLOR_RED, curses.COLOR_BLACK)
            # curses.init_pair(6, curses.COLOR_BLUE, curses.COLOR_WHITE)
            # curses.init_pair(7, curses.COLOR_BLUE, curses.COLOR_BLACK)
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


            #entities[color_str+rank[x_coord]+piece_name[start_board[i]]] = \
            #Entity(x_coord, y_coord, pieces[start_board[i]], piece_style)

            board_window.addstr(y_coord, x_coord, pieces[current_piece.upper()]+" ")
            #board_window.addch(y_coord, x_coord, 'x')


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
    legal_moves = []
    legal_move_str = ""
    san_move_str = ""
    for move in board.legal_moves:
        legal_moves.append(board.san(move))
        legal_moves.append(chess.Move.uci(move))
        legal_moves.append('e4')
        movo_str = chess.Move.uci(move)
        legal_move_str += movo_str + " "
        piece_char = board.piece_at( file[movo_str[0]] + (int(movo_str[1])- 1)\
         * 8 ).symbol()
        if piece_char.upper() == "P":
            piece_char = ""
        san_move_str += piece_char + movo_str[2:4] + " "

    return legal_moves





if __name__ == "__main__":
    main()


# Centering calculations
# start_x_title = int((width // 2) - (len(title) // 2) - len(title) % 2)
# start_x_subtitle = int((width // 2) - (len(subtitle) // 2) - len(subtitle) % 2)
# start_x_keystr = int((width // 2) - (len(keystr) // 2) - len(keystr) % 2)
# start_y = int((height // 2) - 2)
