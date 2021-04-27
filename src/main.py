import sys,os,traceback,random,curses,chess

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
    #'p': '♟︎ ',
    'p': '♙ ',

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


def draw_board(board_window, board_FEN):
    height, width = board_window.getmaxyx()

    # 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR'
    x_coord = 4 #increment by 2
    y_coord = 4 #increment by 2
    og_xcoord = x_coord
    for i in range(len(board_FEN)):
        current_piece = board_FEN[i]

        if current_piece == '/':
            x_coord = og_xcoord
            y_coord += 2
            continue
        elif current_piece.isdigit():
            for j in range(int(current_piece)):
                board_window.addstr(y_coord, x_coord, '.')
                x_coord += 2
            continue
        elif not current_piece.isdigit():
            if current_piece.isupper():
            #    piece_style = style.WHITE
                color_str = "White"
            else:
             #   piece_style = style.BLACK
                color_str = "Black"
            #entities[color_str+rank[x_coord]+piece_name[start_board[i]]] = \
            #Entity(x_coord, y_coord, pieces[start_board[i]], piece_style)

            board_window.addstr(y_coord, x_coord, pieces[current_piece])

            x_coord += 2
            continue
        else:
            print("error parsing starting FEN")
            break

def game_logic(board_window):
    #draw board
    draw_board(board_window, chess.STARTING_BOARD_FEN)   

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
    curses.start_color()
    #curses.use_default_colors()
    curses.init_pair(1, curses.COLOR_CYAN, curses.COLOR_BLACK)
    curses.init_pair(2, curses.COLOR_RED, curses.COLOR_BLACK)
    curses.init_pair(3, curses.COLOR_BLACK, curses.COLOR_WHITE)

    #start windows
    
    board_window = curses.newwin(height-1, width//2, 0, 0)
    info_window = curses.newwin(height//2, width//2, 0, width//2)
    prompt_window = curses.newwin((height-1)//2, width//2, height//2, width//2)

    windows_array = [board_window, info_window, prompt_window]
    

    # Loop where k is the last character pressed
    while (key != ord('q')): # while not quitting

        # Initialization
        stdscr.clear()



        # board_window.border()
        # info_window.border()
        # prompt_window.border()

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
        statusbarstr = "Press 'q' to exit | CHESS-CLI | Pos: {}, {}".format(cursor_x, cursor_y)
        
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
        stdscr.addstr(height-1, 0, statusbarstr)
        stdscr.addstr(height-1, len(statusbarstr), " " * (width - len(statusbarstr) - 1))
        stdscr.attroff(curses.color_pair(3))

        # Turning on attributes for title
        for i in range(len(windows_array)):
            windows_array[i].border()
            windows_array[i].attron(curses.color_pair(2))
            windows_array[i].attron(curses.A_BOLD)

        # Rendering title
        board_window.addstr(0, 0, board_title)
        info_window.addstr(0, 0, info_title)
        prompt_window.addstr(0, 0, prompt_title)

        # Turning off attributes for title
        for i in range(len(windows_array)):
            windows_array[i].attroff(curses.color_pair(2))
            windows_array[i].attroff(curses.A_BOLD)


        game_logic(board_window)


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
