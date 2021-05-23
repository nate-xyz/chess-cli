import sys, os, traceback, random, curses, chess, math, enum, itertools

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

def update_input(prompt_window, key, prompt_x_coord, prompt_y_coord, user_input_string, inputted_str, entered_move, status_str):
    #global prompt_x_coord, prompt_y_coord, user_input_string, inputted_str, entered_move, status_str
    height, width = prompt_window.getmaxyx()

    if key == curses.KEY_MOUSE: #dont do any input for mouse event
        return (prompt_x_coord, prompt_y_coord, user_input_string, inputted_str, entered_move, status_str)
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
        return (prompt_x_coord, prompt_y_coord, user_input_string, inputted_str, entered_move, status_str)
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

    return (prompt_x_coord, prompt_y_coord, user_input_string, inputted_str, entered_move, status_str)


# dP                                        dP              oo                              dP   
# 88                                        88                                              88   
# 88d888b. .d8888b. .d8888b. 88d888b. .d888b88              dP 88d888b. 88d888b. dP    dP d8888P 
# 88'  `88 88'  `88 88'  `88 88'  `88 88'  `88              88 88'  `88 88'  `88 88    88   88   
# 88.  .88 88.  .88 88.  .88 88       88.  .88              88 88    88 88.  .88 88.  .88   88   
# 88Y8888' `88888P' `88888P8 dP       `88888P8              dP dP    dP 88Y888P' `88888P'   dP   
#                                              oooooooooooo             88                       
#                                                                       dP                       
def board_input(screen, key, screen_width, screen_height, board_square_coord, mouse_pressed, floating_piece, floating):
    #global board_square_coord, mouse_pressed, floating_piece, floating
    height, width = screen.getmaxyx()

    if key != curses.KEY_MOUSE: #input needs to be mouse input
        print("1")
        return (mouse_pressed, floating_piece, floating)
    
    try:
        _, mouse_x, mouse_y, _, button_state =  curses.getmouse()
        bs_str = "none"
        if button_state & curses.BUTTON1_PRESSED != 0:
            bs_str = "b1 pressed"
            mouse_pressed = True
        if button_state & curses.BUTTON1_RELEASED != 0:
            bs_str = "b1 released"
            mouse_pressed = False
            floating = False
    
        screen.addstr(2, 2, "mouse_x: {} mouse_y: {} button_state: {}".format( str(mouse_x), str(mouse_y), bs_str))
        key_tuple = (mouse_x, mouse_y)
        if key_tuple in board_square_coord.keys() and mouse_pressed:
            screen.addstr(6, 2, "has key")
            piece_str = board_square_coord[key_tuple][1]
            if piece_str != None and not floating:
                floating = True
                floating_piece = board_square_coord[key_tuple]
                screen.addstr(5, 2, "piece is {}".format(piece_str ))
        if mouse_pressed:
            color_pair = floating_piece[0]
            screen.attron(curses.color_pair(color_pair))
            screen.attron(curses.A_BOLD)
            screen.addstr(mouse_y, mouse_x, floating_piece[1]+" ")
            screen.attron(curses.color_pair(color_pair))
            screen.attron(curses.A_BOLD)
        print("2")
        return (mouse_pressed, floating_piece, floating)
    except:
        screen.addstr(7, 2, "error")
        print("3")
        return (mouse_pressed, floating_piece, floating)