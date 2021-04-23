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
    // let king = 'K';
    // let queen = 'Q';
    // let rook = 'R';
    // let bishop = 'B';
    // let knight = 'K';
    // let pawn = 'p';
    let king = '♚';
    let queen = '♛';
    let rook = '♜';
    let bishop = '♝';
    let knight = '♞';
    let pawn = '♙';
    let piece_list: [char; 6] = ['♚','♛','♜','♝','♞','♙']; 

    //initialize engine
    let mut engine = console_engine::ConsoleEngine::init_fill_require(48, 48, 30);

    let gen_board_x: [i32; 8] = [12, 14, 16, 18, 20, 22, 24, 26];
    let gen_board_y: [i32; 8] = [8, 9, 10, 11, 12, 13, 14, 15];
   
    let dark_square = console_engine::Color::Rgb{r: 209, g: 139, b: 70};
    let light_square = console_engine::Color::Rgb{r: 254, g: 206, b: 158};


    let mut piece_x = 8;
    let mut piece_y = 3;
    let piece_w = 16;
    let piece_h = 4;
    let mut dragging = false;
    let mut relative_x = 0;
    let mut relative_y = 0;

    loop {
        engine.wait_frame(); // wait for next frame + capture inputs
        engine.check_resize(); // resize the terminal if its size has changed
        if engine.is_key_pressed(KeyCode::Char('q')) { // if the user presses 'q' :
        break; // exits app
        }
        engine.clear_screen(); // reset the screen

        

        engine.print(0,4,"CHESS MOMENT!");
        engine.print(1,5,"CHESS MOMENT!");
        engine.print(2,6,"CHESS MOMENT!");

        //alg notation 
        engine.print(gen_board_x[0],gen_board_y[0]-2,"a b c d e f g h");
        engine.print(gen_board_x[0],gen_board_y[7]+2,"a b c d e f g h");
        engine.print(gen_board_x[0]-4,gen_board_y[0],"8\n7\n6\n5\n4\n3\n2\n1");
        engine.print(gen_board_x[7]+4,gen_board_y[0],"8\n7\n6\n5\n4\n3\n2\n1");
        

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
            let mut piece_color = Color::DarkGrey;
            let mut top_row = 0;
            let mut pawn_row = 0;
            if i == 1 {
                top_row = 7;
                pawn_row = 5;
                piece_color = Color::White;
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


        let mouse_pos = engine.get_mouse_press(MouseButton::Left);
        if let Some(mouse_pos) = mouse_pos {
            // if the mouse position is within the boundaries of a piece
            // enables dragging mode and register relative position of the mouse
            // if mouse_pos.0 as i32 >= piece_x
            //     && mouse_pos.0 as i32 <= piece_x + piece_w
            //     && mouse_pos.1 as i32 >= piece_y
            //     && mouse_pos.1 as i32 <= piece_y + piece_h
            // {
            //     dragging = true;
            //     relative_x = mouse_pos.0 as i32 - piece_x;
            //     relative_y = mouse_pos.1 as i32 - piece_y;
            // }


            if piece_list.contains(&engine.get_pxl(mouse_pos.0 as i32, mouse_pos.1 as i32).unwrap().chr) {

                dragging = true;
                engine.print(10,5,"drag");

            }
        }

        // // check if a mouse button is currently held
        // let mouse_pos = engine.get_mouse_held(MouseButton::Left);
        // if let Some(mouse_pos) = mouse_pos {
        //     // if dragging mode is enabled, move the rectangle according to mouse's position
        //     if dragging {
        //         piece_x = mouse_pos.0 as i32 - relative_x;
        //         piece_y = mouse_pos.1 as i32 - relative_y;
        //     }
        // }

        // // check if the mouse has been released
        // let mouse_pos = engine.get_mouse_released(MouseButton::Left);
        // if mouse_pos.is_some() {
        //     // disable dragging mode
        //     dragging = false;
        // }

        // // print the recrangle
        // engine.rect(
        //     piece_x,
        //     piece_y,
        //     piece_x + piece_w,
        //     piece_y + piece_h,
        //     pixel::pxl('#'),
        // );
        // engine.print(piece_x + 4, piece_y + 2, "Drag me!");


        // draw_game(
        //     &mut engine,
        //     &board,
        //     white_pawn
        // );
        engine.draw(); // draw the screen
    }
}
