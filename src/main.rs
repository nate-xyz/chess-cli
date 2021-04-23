//use console_engine::screen::Screen;
use console_engine::pixel;
use console_engine::Color;
use console_engine::{KeyCode, MouseButton};


fn main() {

    //Pieces
    // let white_king = '♔';
    // let white_queen = '♕';
    // let white_rook = '♖';
    // let white_bishop = '♗';
    // let white_knight = '♘';
    //let white_pawn = '♙';
    // let black_king = '♚';
    // let black_queen = '♛';
    // let black_rook = '♜';
    // let black_bishop = '♝';
    // let black_knight = '♞';
    // let black_pawn = "♟︎";
    //let square = '█';
    let king = 'K';
    let queen = 'Q';
    let rook = 'R';
    let bishop = 'B';
    let knight = 'K';
    let pawn = 'p';

    //initialize engine
    let mut engine = console_engine::ConsoleEngine::init_fill_require(48, 48, 30);

    //initialize board screen
    //let mut board = Screen::new(24, 24);

    //let gen_board: [i32; 8] = [0, 3, 6, 9, 12, 15, 18, 21];
    //let gen_board: [i32; 8] = [12, 15, 18, 21, 24, 27, 30, 33];
    let gen_board_x: [i32; 8] = [12, 14, 16, 18, 20, 22, 24, 26];
    let gen_board_y: [i32; 8] = [8, 9, 10, 11, 12, 13, 14, 15];
   
    let dark_square = Color::White;
    let light_square = Color::DarkGrey;

    loop {
        engine.wait_frame(); // wait for next frame + capture inputs
        engine.check_resize(); // resize the terminal if its size has changed
        if engine.is_key_pressed(KeyCode::Char('q')) { // if the user presses 'q' :
        break; // exits app
        }
        engine.clear_screen(); // reset the screen

        let mouse_pos = engine.get_mouse_press(MouseButton::Left);

       
        engine.print(0,4,"CHESS MOMENT!");

        //draw board
        for y in &gen_board_y {
            for x in &gen_board_x {
                let square_color = if  [12, 16, 20, 24].contains(x) { if *y % 2 == 0 {dark_square} else {light_square} } else  { if *y % 2 == 0 {light_square} else {dark_square} };
                //engine.fill_rect(*x, *y, *x+2, *y+2, pixel::pxl_fg('x', square_color));
                engine.set_pxl(*x, *y, pixel::pxl_fg('x', square_color));
            }
        }

        //draw pieces
        for i in 0..=1 {
            let mut piece_color = Color::Blue;
            let mut top_row = 0;
            let mut pawn_row = 0;
            if i == 1 {
                top_row = 7;
                pawn_row = 5;
                piece_color = Color::Red;
            }
            engine.set_pxl(gen_board_x[0], gen_board_y[0]+top_row, pixel::pxl_fg(rook, piece_color));
            engine.set_pxl(gen_board_x[1], gen_board_y[0]+top_row, pixel::pxl_fg(knight, piece_color));
            engine.set_pxl(gen_board_x[2], gen_board_y[0]+top_row, pixel::pxl_fg(bishop, piece_color));
            engine.set_pxl(gen_board_x[3], gen_board_y[0]+top_row, pixel::pxl_fg(queen, piece_color));
            engine.set_pxl(gen_board_x[4], gen_board_y[0]+top_row, pixel::pxl_fg(king, piece_color));
            engine.set_pxl(gen_board_x[5], gen_board_y[0]+top_row, pixel::pxl_fg(bishop, piece_color));
            engine.set_pxl(gen_board_x[6], gen_board_y[0]+top_row, pixel::pxl_fg(knight, piece_color));
            engine.set_pxl(gen_board_x[7], gen_board_y[0]+top_row, pixel::pxl_fg(rook, piece_color));

            engine.set_pxl(gen_board_x[0], gen_board_y[1]+pawn_row, pixel::pxl_fg(pawn, piece_color));
            engine.set_pxl(gen_board_x[1], gen_board_y[1]+pawn_row, pixel::pxl_fg(pawn, piece_color));
            engine.set_pxl(gen_board_x[2], gen_board_y[1]+pawn_row, pixel::pxl_fg(pawn, piece_color));
            engine.set_pxl(gen_board_x[3], gen_board_y[1]+pawn_row, pixel::pxl_fg(pawn, piece_color));
            engine.set_pxl(gen_board_x[4], gen_board_y[1]+pawn_row, pixel::pxl_fg(pawn, piece_color));
            engine.set_pxl(gen_board_x[5], gen_board_y[1]+pawn_row, pixel::pxl_fg(pawn, piece_color));
            engine.set_pxl(gen_board_x[6], gen_board_y[1]+pawn_row, pixel::pxl_fg(pawn, piece_color));
            engine.set_pxl(gen_board_x[7], gen_board_y[1]+pawn_row, pixel::pxl_fg(pawn, piece_color));
        }


        

        // draw_game(
        //     &mut engine,
        //     &board,
        //     white_pawn
        // );
        engine.draw(); // draw the screen
    }
}
