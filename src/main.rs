extern crate ncurses;
extern crate libm;
extern crate chess;

use libm::floor;
use ncurses::*;

fn main() {
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

        for window in windows_array.iter() {
            box_(*window, 0, 0);
            wrefresh(*window);

        }
        
        key = getch()
    }
    endwin();

//chess display
fn chess_display() {

    //let (mut height, mut width): (f64, f64) = info_window.getmaxyx();


}
}