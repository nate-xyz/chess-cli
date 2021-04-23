use console_engine::pixel;
use console_engine::Color;
use console_engine::KeyCode;




fn main() {
    let square = '█';

    // initializes a screen of 20x10 characters with a target of 3 frames per second
    // coordinates will range from [0,0] to [19,9]
    let mut engine = console_engine::ConsoleEngine::init(40, 20, 10);
    //let value = 14;
    // main loop, be aware that you'll have to break it because ctrl+C is captured

    let mut square_x = 0;
    let mut square_y = 0;

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
        for mut y in 0..=7 {
            for mut x in 0..=7 {
 
                if x % 2 == 0 {
                    if y % 2 == 0 {
                        engine.rect(x, y, x, y, pixel::pxl_fg(square, Color::Black));
                    } 
                    else {
                        engine.rect(x, y, x, y, pixel::pxl_fg(square, Color::Grey));
                    }
                }
                else {
                    if y % 2 == 0 {
                        engine.rect(x, y, x, y, pixel::pxl_fg(square, Color::Grey));
                    } 
                    else {
                        engine.rect(x, y, x, y, pixel::pxl_fg(square, Color::Black));
                    }
                }

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
