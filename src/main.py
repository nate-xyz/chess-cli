import sys, os, traceback, random, curses, chess


board = chess.Board()

#prompt vars
prompt_x_coord = 1
prompt_y_coord = 1
user_input_string = ""
inputted_str = ""
last_move_str = "no move yet"
status_str = ""
legal_move_str = ""
san_move_str = ""
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


def display_info(info_window):
    #TODO: text wrapping to avoid errors when the terminal is too small
    global last_move_str
    global status_str
    global inputted_str
    global legal_move_str
    global san_move_str
    #san_move_str = legal_move_str[2:4]
    height, width = info_window.getmaxyx()\


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
        info_window.attron(curses.color_pair(8))
    else:
        info_window.attron(curses.color_pair(9))
    info_window.addstr(3, 1, status_str)
    info_window.attroff(curses.color_pair(9))
    info_window.addstr(4, 1, "{}: {}".format("input",inputted_str))
    info_window.attron(curses.color_pair(8))
    info_window.addstr(5, 1, "{}: {}".format("legal moves", san_move_str))
    info_window.attroff(curses.color_pair(9))

    status_str = ""

def update_input(prompt_window, key):
    global prompt_x_coord
    global prompt_y_coord
    global user_input_string
    global inputted_str
    global entered_move
    height, width = prompt_window.getmaxyx()

    prompt_window.addch(prompt_y_coord, prompt_x_coord, key)
    prompt_window.addch(prompt_y_coord, 0, '>')

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


def game_logic(board_window):
    global inputted_str
    global board
    global status_str
    global entered_move
    global last_move_str
    inputted_str = inputted_str.strip(' ').strip('\0').strip('^@')
    legal_moves = []

    if entered_move:
        entered_move = False
        if inputted_str == 'undo':
            board.pop()
        else:
            if inputted_str not in legal_moves:
                status_str = "last input is invalid"
            else:
                status_str = "move is legal!"
                if board.is_legal(chess.Move.from_uci(inputted_str)):
                    last_move_str = chess.Move.from_uci(inputted_str)
                    curses.flash()
                    curses.beep()
                    board.push(chess.Move.from_uci(inputted_str))
    #draw board
    draw_board(board_window, board.board_fen())
    legal_moves = generate_legal_moves()


def generate_legal_moves():
    global legal_move_str
    global san_move_str
    global board
    legal_moves = []
    legal_move_str = ""
    san_move_str = ""
    for move in board.legal_moves:
        legal_moves.append(chess.Move.uci(move))
        movo_str = chess.Move.uci(move)
        legal_move_str += movo_str + " "
        piece_char = board.piece_at( file[movo_str[0]] + (int(movo_str[1])- 1)\
         * 8 ).symbol()
        if piece_char.upper() == "P":
            piece_char = ""
        san_move_str += piece_char + movo_str[2:4] + " "
    return legal_moves


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

    board_window = curses.newwin(height-1, width//2, 0, 0)
    info_window = curses.newwin(height//2, width//2, 0, width//2)
    prompt_window = curses.newwin((height-1)//2, width//2, height//2, width//2)

    windows_array = [board_window, info_window, prompt_window]

    # Loop where k is the last character pressed
    while (key != 15): # while not quitting

        # Initialization
        stdscr.clear()
        board_window.clear()
        info_window.clear()

        #resize everything if necessary
        if curses.is_term_resized(height, width):
            height, width = stdscr.getmaxyx()
            stdscr.clear()
            curses.resize_term(height, width)
            stdscr.refresh()

            board_window.resize(height-1, width//2)
            info_window.resize(height//2, width//2)
            prompt_window.resize((height-1)//2, width//2)

            board_window.mvwin(0, 0)
            info_window.mvwin(0, width//2)
            prompt_window.mvwin(height//2, width//2)

            board_window.clear()
            info_window.clear()
            prompt_window.clear()
            board_window.refresh()
            info_window.refresh()
            prompt_window.refresh()

        #get winodw dimensions
        height, width = stdscr.getmaxyx()
        board_window_height, board_window_width = board_window.getmaxyx()
        info_window_height, info_window_width = info_window.getmaxyx()
        prompt_window_height, prompt_window_width = prompt_window.getmaxyx()

        cursor_x = max(0, cursor_x)
        cursor_x = min(width-1, cursor_x)

        cursor_y = max(0, cursor_y)
        cursor_y = min(height-1, cursor_y)

        # Declaration of strings
        board_title = "board"[:width-1]
        info_title = "info"[:width-1]
        prompt_title = "prompt"[:width-1]

        keystr = "Last key pressed: {}".format(key)[:width-1]
        #statusbarstr = "Press 'Ctrl+o' to exit | CHESS-CLI | Pos: {}, {}".format(cursor_x, cursor_y)
        statusbarstr = "Press 'Ctrl+o' to exit | CHESS-CLI"


        statusbarfull = "{} | {}".format(statusbarstr, keystr)
        #statusbarfull = ""

        if key == 0:
            keystr = "No key press detected..."[:width-1]

        # Centering calculations
        # start_x_title = int((width // 2) - (len(title) // 2) - len(title) % 2)
        # start_x_subtitle = int((width // 2) - (len(subtitle) // 2) - len(subtitle) % 2)
        # start_x_keystr = int((width // 2) - (len(keystr) // 2) - len(keystr) % 2)
        # start_y = int((height // 2) - 2)

        #render strings
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
        board_window.refresh()
        info_window.refresh()
        prompt_window.refresh()

        # Wait for next input
        key = stdscr.getch()


def main():
    curses.wrapper(draw_screen)


if __name__ == "__main__":
    main()
