use console_engine::pixel;
use console_engine::Color;
use console_engine::KeyCode;




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

    // initializes a screen of 20x10 characters with a target of 3 frames per second
    // coordinates will range from [0,0] to [19,9]
    let mut engine = console_engine::ConsoleEngine::init(40, 40, 10);
    //let value = 14;
    // main loop, be aware that you'll have to break it because ctrl+C is captured

    //let mut square_x = 0;
    //let mut square_y = 0;

    loop {
        engine.wait_frame(); // wait for next frame + capture inputs
        engine.clear_screen(); // reset the screen



        // pub fn rect(
        //     &mut self,
        //     start_x: i32,
        //     start_y: i32,
        //     end_x: i32,
        //     end_y: i32,
        //     character: Pixel
        // )

        //draw board
        let gen_board: [i32; 8] = [0, 3, 6, 9, 12, 15, 18, 21];
        let dark_square = Color::Black;
        let light_square = Color::Grey;

        for y in &gen_board {
            for x in &gen_board {
                let square_color = if *x % 2 == 0 { if *y % 2 == 0 {dark_square} else {light_square} } else  { if *y % 2 == 0 {light_square} else {dark_square} };



                //engine.rect(*x, *y, *x, *y, pixel::pxl_fg(square, square_color));
                // engine.rect(*x+1, *y, *x+1, *y, pixel::pxl_fg(square, square_color));
                // engine.rect(*x+2, *y, *x+2, *y, pixel::pxl_fg(square, square_color));

                engine.fill_rect(*x, *y, *x+2, *y+2, pixel::pxl_fg(square, square_color));


            }
        }
        //engine.rect(0, 0, 0, 0, pixel::pxl_fg('█', Color::Cyan));



        //engine.line(0, 0, 19, 9, pixel::pxl('█')); // draw a line of '#' from [0,0] to [19,9]
        //engine.print(0, 4, format!("Result: {}", value).as_str()); // prints some value at [0,4]

        //engine.set_pxl(4, 0, pixel::pxl_fg('O', Color::Cyan)); // write a majestic cyan 'O' at [4,0]

        if engine.is_key_pressed(KeyCode::Char('q')) { // if the user presses 'q' :
            break; // exits app
        }

        engine.draw(); // draw the screen
    }
}
