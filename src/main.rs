#![forbid(unsafe_code)]


extern crate ncurses;
extern crate libm;
extern crate chess;

use chess::Board;
use libm::floor;
use ncurses::*;
use std::*;
use std::collections::HashMap;
use std::iter::Iterator;

fn main() {

    fn Pieces_hash() {
         let mut pieces_hash = HashMap::new();
         pieces_hash.insert("K", vec!["♔"]);
         pieces_hash.insert("Q", vec!["♕"]);
         pieces_hash.insert("R", vec!["♖"]);
         pieces_hash.insert("B", vec!["♗"]);
         pieces_hash.insert("N", vec!["♘"]);
         pieces_hash.insert("P", vec!["♙"]);
         pieces_hash.insert("k", vec!["♚"]);
         pieces_hash.insert("q", vec!["♛"]);
         pieces_hash.insert("r", vec!["♜"]);
         pieces_hash.insert("b", vec!["♝"]);
         pieces_hash.insert("n", vec!["♞"]);
         //pieces_hash.insert("p", vec!["♟︎"]);
         pieces_hash.insert("p", vec!["♙"]);
    }


    let mut board = Board::default();


    initscr();

    let mut height = 0; 
    let mut width = 0; 
    getmaxyx(stdscr(), &mut height, &mut width);

    let mut key = 0;

    // Clear and refresh the screen for a blank canvas
    clear();
    refresh();
    keypad(stdscr(), true);
    noecho();

    // allow input, Start colors in curses
    //echo()
    curs_set(CURSOR_VISIBILITY::CURSOR_VISIBLE);
    start_color();

    init_pair(1, COLOR_CYAN, COLOR_BLACK);
    init_pair(2, COLOR_RED, COLOR_BLACK);
    init_pair(3, COLOR_BLACK, COLOR_WHITE);

    let light_square = 215; //SandyBrown
    let dark_square = 94; //Orange4
    let light_piece = 230; //Cornsilk1
    let dark_piece = 233; //Grey7

    init_pair(4, light_piece, light_square);
    init_pair(5, light_piece, dark_square);
    init_pair(6, dark_piece, light_square);
    init_pair(7, dark_piece, dark_square);

    //is_floating_bool piece colors
    init_pair(10, light_piece, dark_piece);
    init_pair(11, dark_piece, light_piece);

    //move legality colors
    init_pair(8, COLOR_BLACK, COLOR_GREEN);
    init_pair(9, COLOR_WHITE, COLOR_RED);

    //start windows
    //newwin(lines: i32, cols: i32, y: i32, x: i32)

    let mut f_height = height as f64; 
    let mut f_width = width as f64; 

    let board_window = newwin( floor((f_height/4.0)*3.0) as i32, floor(f_width/2.0) as i32, 0, 0);
    let prompt_window = newwin( floor((f_height/4.0)-1.0) as i32, floor(f_width/2.0) as i32,  floor((f_height/4.0)*3.0) as i32, 0);
    let info_window = newwin( floor(f_height/2.0) as i32, floor(f_width/2.0) as i32, 0, floor(f_width/2.0) as i32);
    let history_window = newwin( floor((f_height/2.0)-1.0) as i32, floor(f_width/2.0) as i32, floor(f_height/2.0) as i32, floor(f_width/2.0) as i32);

    let windows_array = vec![board_window, info_window, prompt_window, history_window]; 

    //main loop, exits when (ctrl+o) is pressed
    while key != 15 {

        //resize everything if necessary
        if is_term_resized(height, width) {
            getmaxyx(stdscr(), &mut height, &mut width);
            resize_term(height, width);

            clear();
            refresh();

            for window in windows_array.iter() {
                box_(*window, 0, 0);
                wrefresh(*window);
    
            }


        }

        //python old code


        //get window dimensions

        //height, width = stdscr.getmaxyx()
        // board_window_height, board_window_width = board_window.getmaxyx()
        // info_window_height, info_window_width = info_window.getmaxyx()
        // prompt_window_height, prompt_window_width = prompt_window.getmaxyx()
        // history_window_height, history_window_width = history_window.getmaxyx()

        let mut board_window_height = 0; 
        let mut info_window_height= 0; 
        let mut prompt_window_height = 0; 
        let mut history_window_height = 0; 

        let mut board_window_width = 0; 
        let mut info_window_width = 0; 
        let mut prompt_window_width = 0; 
        let mut history_window_width = 0; 
 
        getmaxyx(board_window, &mut board_window_height, &mut board_window_width);
        getmaxyx(info_window, &mut info_window_height, &mut info_window_width);
        getmaxyx(prompt_window, &mut prompt_window_height, &mut prompt_window_width);
        getmaxyx(history_window, &mut history_window_height, &mut history_window_width);


        //Declaration of strings

        // board_title = "board"[:width-1]
        // info_title = "info"[:width-1]
        // prompt_title = "prompt"[:width-1]
        // history_title = "move_history"[:width-1]

        let board_title = "board";
        let info_title = "info";
        let prompt_title = "prompt";
        let history_title = "move_history";

        let keystr = format!("Last key pressed: {}", key);
        let statusbarstr = "Press 'Ctrl+o' to exit | CHESS-CLI";
        let statusbarfull = format!("{} | {}", statusbarstr, keystr);

        if key == 0 {
            let keystr = "No key press detected...";
        }
            

        //Render status bar
        // stdscr.attron(curses.color_pair(3))
        // stdscr.addstr(height-1, 0, statusbarfull)
        // stdscr.addstr(height-1, len(statusbarfull), " " * (width - len(statusbarfull) - 1))
        // stdscr.attroff(curses.color_pair(3))

        attr_on(COLOR_PAIR(3));
        //function syntax: mvwaddstr(w: WINDOW, y: i32, x: i32, s: &str) -> i32
        mvaddstr(height-1, 0, &statusbarfull);
        mvaddstr(height-1, statusbarfull.chars().count() as i32, &" ".repeat((width - statusbarfull.chars().count() as i32 - 1) as usize) );
        attroff(COLOR_PAIR(3));

        for window in windows_array.iter() {
            box_(*window, 0, 0);
            wrefresh(*window);
        }

// ===================================================================================================================================================
// ===================================================  =========    ===================================================================  ==  ========
// ===================================================  ========  ==  ==================================================================  ==  ========
// ================  =================================  ========  ============================  ========================================  ==  ========
// ==   ===  =  ==    ===   ===  =   ===  = ====   ===  =======    =====  =  ==  = ====   ===    ==  ===   ===  = =========   ====   ===  ==  ===   ==
// =  =  ==  =  ===  ===  =  ==    =  ==     ==  =  ==  ========  ======  =  ==     ==  =  ===  =======     ==     =======  =  ==  =  ==  ==  ==  =  =
// =     ===   ====  ===     ==  =======  =  =====  ==  ========  ======  =  ==  =  ==  ======  ===  ==  =  ==  =  =======  ========  ==  ==  ===  ===
// =  ======   ====  ===  =====  =======  =  ===    ==  ========  ======  =  ==  =  ==  ======  ===  ==  =  ==  =  =======  ======    ==  ==  ====  ==
// =  =  ==  =  ===  ===  =  ==  =======  =  ==  =  ==  ========  ======  =  ==  =  ==  =  ===  ===  ==  =  ==  =  =======  =  ==  =  ==  ==  ==  =  =
// ==   ===  =  ===   ===   ===  =======  =  ===    ==  ========  =======    ==  =  ===   ====   ==  ===   ===  =  ========   ====    ==  ==  ===   ==
// ===================================================================================================================================================
                                                                                                                       
                                                                                                         

        // #update_input updates the game screen prompt window and returns what the user is currently typing
        // input_buffer_str, move_str, entered_move_bool, status_str = \
        //     update_input(prompt_window, key, input_buffer_str, move_str, \
        // entered_move_bool, status_str)
        
        // #update the board window mouse input
        // #mouse_pressed_bool, floating_piece_str, is_floating_bool = board_window_mouse_input(board_window, key, width, height, board_square_coord, mouse_pressed_bool, floating_piece_str, is_floating_bool)
       
        // #game_logic determines if an inputted move is legal and manages the gamestate
        // if not ai_game_bool:
        //     #call local_game_logic
        //     move_str, board, status_str, entered_move_bool, last_move_str, \
        //     history_arr, move_amount, \
        //     post_screen_toggle, board_square_coord, legal_move_str, san_move_str = \
        //         local_game_logic(board_window, move_str, board, \
        //     status_str, entered_move_bool, history_arr, \
        //      move_amount, \
        //     post_screen_toggle, board_square_coord, pieces, \
        //     legal_move_str, san_move_str, outcome_tuple)
        // else:
        //     #call play_stockfish
        //     move_str, board, status_str, entered_move_bool, last_move_str, \
        //     history_arr, move_amount, \
        //     post_screen_toggle, board_square_coord, legal_move_str, san_move_str, stockfish_obj = \
        //         stockfish_logic( board_window, move_str, board, 
        //     status_str, entered_move_bool, history_arr, \
        //     move_amount, \
        //     post_screen_toggle, board_square_coord, pieces, \
        //     legal_move_str, san_move_str, outcome_tuple, stockfish_obj)
        
        // if post_screen_toggle: #check if post_screen is enabled
        //     post_screen_toggle = False

        //     #post_screen displays after the win condition has been met
        //     input_buffer_str, move_str, entered_move_bool, history_arr, status_str \
        //          = post_screen(stdscr, input_buffer_str, move_str, entered_move_bool, history_arr, \
        //             status_str, board_square_coord, pieces)
        //     if quit_game_bool:
        //         break

        //     #return to the welcome screen
        //     input_buffer_str, move_str, entered_move_bool, status_str = \
        //         welcome_screen(stdscr, input_buffer_str, move_str, entered_move_bool, status_str)
        //     continue
        
        // #windows for the game screen

        // #display game information
        // status_str, legal_move_str, san_move_str = \
        //     display_info(board, info_window, last_move_str, status_str, move_str, legal_move_str, san_move_str)
        // #display move history
        // display_history(history_window, history_arr, move_amount, pieces)

        // #end of external function call section 


// =======================================================================================================================================
// ==================  ================    ==========================================================  =====================  ==  ========
// ==================  ===============  ==  =========================================================  =====================  ==  ========
// ==================  ===============  ==========================  =================================  =====================  ==  ========
// ==   ===  = ======  ========   ===    ===========   ===  =  ==    ===   ===  =   ===  = ====   ===  ========   ====   ===  ==  ===   ==
// =  =  ==     ===    =======     ===  ===========  =  ==  =  ===  ===  =  ==    =  ==     ==  =  ==  =======  =  ==  =  ==  ==  ==  =  =
// =     ==  =  ==  =  =======  =  ===  ===========     ===   ====  ===     ==  =======  =  =====  ==  =======  ========  ==  ==  ===  ===
// =  =====  =  ==  =  =======  =  ===  ===========  ======   ====  ===  =====  =======  =  ===    ==  =======  ======    ==  ==  ====  ==
// =  =  ==  =  ==  =  =======  =  ===  ===========  =  ==  =  ===  ===  =  ==  =======  =  ==  =  ==  =======  =  ==  =  ==  ==  ==  =  =
// ==   ===  =  ===    ========   ====  ============   ===  =  ===   ===   ===  =======  =  ===    ==  ========   ====    ==  ==  ===   ==
// =======================================================================================================================================
                                                                                                                       
        //Turning on attributes for title
        for window in windows_array.iter() {            
            //wattron(w: WINDOW, attr: NCURSES_ATTR_T)
            wattron(*window, COLOR_PAIR(2));
            wattron(*window, A_BOLD());
        }
    
        //Rendering title
        //mvwaddstr(w: WINDOW, y: i32, x: i32, s: &str) -> i32
        mvwaddstr(board_window, 0, 1, board_title); 
        mvwaddstr(info_window, 0, 1, info_title);  
        mvwaddstr(prompt_window, 0, 1, prompt_title); 
        mvwaddstr(history_window, 0, 1, history_title); 

        // Turning off attributes for title
        for window in windows_array.iter() {            
            wattroff(*window, COLOR_PAIR(2));
            wattroff(*window, A_BOLD());
        }
        
        // Refresh the screen
        refresh();
        for window in windows_array.iter() {
            wrefresh(*window);
        }
        
        key = getch()
    }
    endwin();

//chess display
//fn chess_display(board, info_window: i32, last_move_str: str, status_str, move_str, legal_move_str, san_move_str) {

//    getmaxyx(info_window, &mut info_window_height, &mut info_window_width);


//}
}