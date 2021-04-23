use console_engine::screen::Screen;
use console_engine::pixel;
use console_engine::Color;
use console_engine::KeyCode;



fn draw_game(
    engine: &mut console_engine::ConsoleEngine,
    game_scr: &Screen,
    piece: char

) {
    // print the game screen at specific coordinates
    engine.print_screen(12, 12, &game_scr);
    engine.print_screen_alpha(1, 1, &game_scr, piece);


}


fn main() {

    //Pieces
    let white_king = '♔';
    let white_queen = '♕';
    let white_rook = '♖';
    let white_bishop = '♗';
    let white_knight = '♘';
    let white_pawn = '♙';

    let black_king = '♚';
    let black_queen = '♛';
    let black_rook = '♜';
    let black_bishop = '♝';
    let black_knight = '♞';
    let black_pawn = "♟︎";

    let square = '█';

    //initialize engine
    let mut engine = console_engine::ConsoleEngine::init(48, 48, 30);
    //initialize board screen
    let mut board = Screen::new(24, 24);

    //draw board
    let gen_board: [i32; 8] = [0, 3, 6, 9, 12, 15, 18, 21];
    let dark_square = Color::Black;
    let light_square = Color::DarkGrey;

    for y in &gen_board {
        for x in &gen_board {
            let square_color = if *x % 2 == 0 { if *y % 2 == 0 {dark_square} else {light_square} } else  { if *y % 2 == 0 {light_square} else {dark_square} };
            board.fill_rect(*x, *y, *x+2, *y+2, pixel::pxl_fg(square, square_color));
        }
    }

    loop {
        engine.wait_frame(); // wait for next frame + capture inputs
        if engine.is_key_pressed(KeyCode::Char('q')) { // if the user presses 'q' :
        break; // exits app
        }
        engine.clear_screen(); // reset the screen



        

        draw_game(
            &mut engine,
            &board,
            white_pawn
        );



        engine.draw(); // draw the screen
    }
}
