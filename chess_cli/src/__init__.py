
import chess
import py_cui
#import stockfish

class CHESSCUI:
################################################################################
    def __init__(self, master):

        self.master = master
        self.master.set_status_bar_text('Testing')

        self.positions = []
        for i in range (0,8):
            row = []
            for j in range(0,8):
                pos = self.master.add_button('0', i, j, command=None)
                row.append(pos)
            self.positions.append(row)
        scroll_menu_options = ['New Game', 'Exit']

        self.menu = self.master.add_scroll_menu('Menu', 2, 4, row_span = 2, column_span = 3)
        self.menu.add_item_list(scroll_menu_options)
        self.menu.add_key_command(py_cui.keys.KEY_ENTER, self.operate_on_menu_item)
        self.menu.set_focus_text('Arrows to select menu items, and enter to use a menu item')
        self.initialize_new_game()
################################################################################
    def initialize_new_game(self):

        initial_placement = self.generate_initial_placement()
        self.game_instance = GAME.Game(initial_placement)
        self.apply_board_state()
        self.master.move_focus(self.menu)
################################################################################
    def operate_on_menu_item(self):

        operation = self.menu.get()
        if operation == 'New Game':
            self.initialize_new_game()
        elif operation == 'Exit':
            exit()
################################################################################
    def apply_board_state(self):

        for i in range(0, 8):
            for j in range(0, 8):
                val = self.game_instance.game_board.board_positions[i][j]
                self.positions[i][j].set_title(str(val))
                pp = (val/(val/2))
                if pp == 2:
                    self.positions[i][j].set_color(py_cui.BLACK_ON_WHITE)
                elif pp != 2:
                    self.positions[i][j].set_color(py_cui.BLACK_ON_CYAN)
################################################################################
def generate_initial_placement(self):
        """Function that creates an initial board placement
        Returns
        -------
        initial_placement : list of list of int
            Two random coordinates for the initial values
        """

        initial_placement = []
        placement_1 = []
        placement_2 = []

        for i in range(0,2):
            placement_1.append(random.randint(0,3))
            placement_2.append(random.randint(0,3))

        placement_1.append(random.randint(1,2) * 2)
        placement_2.append(random.randint(1,2) * 2)

        initial_placement.append(placement_1)
        initial_placement.append(placement_2)
        return initial_placement
################################################################################
def main():

    root = py_cui.PyCUI(4,7)
    root.set_title('chess-cli')
    root.toggle_unicode_borders()
    cui_chess = CHESSCUI(root)
    root.start()
