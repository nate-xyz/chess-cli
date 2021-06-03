import sys, os, traceback, random, curses, chess, math, enum, itertools, stockfish

prompt_x_coord = 1
prompt_y_coord = 1
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

#update_input updates the game screen prompt window and returns what the user is currently typing
def update_input(prompt_window, key, \
                 input_buffer_str, move_str, entered_move_bool, status_str):
    #global prompt_x_coord, prompt_y_coord, input_buffer_str, move_str, entered_move_bool, status_str
    global prompt_x_coord, prompt_y_coord
    height, width = prompt_window.getmaxyx()

    #ascii key codes
    enter_key = 10
    space = 32
    octothorpe = 35 # # key
    plus_sign = 43 # + key
    delete_key = 127
    
    input_keys = set((octothorpe, plus_sign, space)) #set of valid input keys that are not alphanumeric

    up_arrow = 259
    down_arrow = 258
    left_arrow = 260
    right_arrow = 261

    if key == curses.KEY_MOUSE: #dont do any input for mouse event
        return (prompt_x_coord, prompt_y_coord, input_buffer_str, \
                move_str, entered_move_bool, status_str)

    if key == delete_key: 
        if prompt_x_coord-1 <= 0:
            delete_x = 1
        else:
            delete_x = prompt_x_coord-1
        prompt_window.addch(prompt_y_coord, delete_x, chr(8248)) #clear last char pointer
        prompt_window.addch(prompt_y_coord, delete_x+1, ' ') #clear last char printed
        prompt_x_coord -= 1 #decrement char position
        input_buffer_str = input_buffer_str[:-1]
    elif chr(key).isalnum() or key in input_keys:
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
        return (prompt_x_coord, prompt_y_coord, input_buffer_str, \
                move_str, entered_move_bool, status_str)
        # for i in range(1, height-1):
        #     prompt_window.addstr(i, prompt_x_coord, " " * (width-1))
    
    if key == enter_key: 
        entered_move_bool = True 
        move_str = input_buffer_str #set global string to check if move is legal
        input_buffer_str = "" #reset input buffer
        prompt_x_coord = 1 #reset char coordinates
        prompt_y_coord = 1#reset char coordinates
        #prompt_window.addch(prompt_y_coord, 0, '|')
        #prompt_window.addch(prompt_y_coord, 0, '>')

        for i in range(1, height-1): #clear window
            prompt_window.addstr(i, prompt_x_coord, " " * (width-1))
    
    #add to the current input buffer
    if key != enter_key and key != delete_key and (chr(key).isalnum() \
              or key in input_keys): #not enter and not delete
        input_buffer_str += chr(key)

    #redraw border in case it was painted over
    prompt_window.border()
    prompt_window.addch(prompt_y_coord, 0, '>') #indicate line youre on

    return ( input_buffer_str, move_str, \
            entered_move_bool, status_str)


# dP                                        dP              oo                              dP   
# 88                                        88                                              88   
# 88d888b. .d8888b. .d8888b. 88d888b. .d888b88              dP 88d888b. 88d888b. dP    dP d8888P 
# 88'  `88 88'  `88 88'  `88 88'  `88 88'  `88              88 88'  `88 88'  `88 88    88   88   
# 88.  .88 88.  .88 88.  .88 88       88.  .88              88 88    88 88.  .88 88.  .88   88   
# 88Y8888' `88888P' `88888P8 dP       `88888P8              dP dP    dP 88Y888P' `88888P'   dP   
#                                              oooooooooooo             88                       
#                                                                       dP                       

# #checks board window for mouse movement and handles mouse input
# def board_window_mouse_input(screen, key, screen_width, screen_height, board_square_coord, mouse_pressed_bool, is_floating_bool_piece_str, is_floating_bool):
#     #global board_square_coord, mouse_pressed_bool, is_floating_bool_piece_str, is_floating_bool
#     height, width = screen.getmaxyx()

#     if key != curses.KEY_MOUSE: #input needs to be mouse input
#         return (mouse_pressed_bool, is_floating_bool_piece_str, is_floating_bool)
    
#     #try except block for getmouse() errors
#     try: 
#         _, mouse_x, mouse_y, _, button_state =  curses.getmouse()
#         bs_str = "none"
        
#         if button_state & curses.BUTTON1_PRESSED != 0:
#             bs_str = "b1 pressed"
#             mouse_pressed_bool = True
        
#         if button_state & curses.BUTTON1_RELEASED != 0:
#             bs_str = "b1 released"
#             mouse_pressed_bool = False
#             is_floating_bool = False
    
#         screen.addstr(2, 2, "mouse_x: {} mouse_y: {} button_state: {}".format( str(mouse_x), str(mouse_y), bs_str))
#         key_tuple = (mouse_x, mouse_y)
        
#         if key_tuple in board_square_coord.keys() and mouse_pressed_bool:
#             screen.addstr(6, 2, "has key")
#             piece_str = board_square_coord[key_tuple][1]
#             if piece_str != None and not is_floating_bool:
#                 is_floating_bool = True
#                 is_floating_bool_piece_str = board_square_coord[key_tuple]
#                 screen.addstr(5, 2, "piece is {}".format(piece_str ))
            
#         if mouse_pressed_bool:
#             color_pair = is_floating_bool_piece_str[0]
#             screen.attron(curses.color_pair(color_pair))
#             screen.attron(curses.A_BOLD)
#             screen.addstr(mouse_y, mouse_x, is_floating_bool_piece_str[1]+" ")
#             screen.attron(curses.color_pair(color_pair))
#             screen.attron(curses.A_BOLD)
#         return (mouse_pressed_bool, is_floating_bool_piece_str, is_floating_bool)

#     except:
#         screen.addstr(7, 2, "error")
#         return (mouse_pressed_bool, is_floating_bool_piece_str, is_floating_bool)