use console_engine::screen::Screen;
use console_engine::pixel;

fn main() {
    // create a screen of 20x11 characters
    let mut scr = Screen::new(20,11);

    // draw some shapes and prints some text
    scr.rect(0,0, 19,10,pixel::pxl('#'));
    scr.fill_circle(5,5, 3, pixel::pxl('*'));
    scr.print(11,4, "Hello,");
    scr.print(11,5, "World!");

    // print the screen to the terminal
    scr.draw();
}
