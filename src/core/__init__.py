import fcntl
import sys
import os
import time
import tty
import termios
os.system("")

# Entities.
class Entity:
  def __init__(self, x, y, char, style):
    self.x = x
    self.y = y
    self.char = char
    self.style = style
  def update(self):
    pass
# Class of different styles
class style():
    BLACK = '\033[30m'
    RED = '\033[31m'
    GREEN = '\033[32m'
    YELLOW = '\033[33m'
    BLUE = '\033[34m'
    MAGENTA = '\033[35m'
    CYAN = '\033[36m'
    WHITE = '\033[37m'
    UNDERLINE = '\033[4m'
    RESET = '\033[0m'

class raw(object):
    def __init__(self, stream):
        self.stream = stream
        self.fd = self.stream.fileno()
    def __enter__(self):
        self.original_stty = termios.tcgetattr(self.stream)
        tty.setcbreak(self.stream)
    def __exit__(self, type, value, traceback):
        termios.tcsetattr(self.stream, termios.TCSANOW, self.original_stty)

class nonblocking(object):
    def __init__(self, stream):
        self.stream = stream
        self.fd = self.stream.fileno()
    def __enter__(self):
        self.orig_fl = fcntl.fcntl(self.fd, fcntl.F_GETFL)
        fcntl.fcntl(self.fd, fcntl.F_SETFL, self.orig_fl | os.O_NONBLOCK)
    def __exit__(self, *args):
        fcntl.fcntl(self.fd, fcntl.F_SETFL, self.orig_fl)




def print_map(map):
  print('\x1bc')
  x = ""
  for i in map:
   y = ""
   for j in i:
     y = y + j
   x = x + y + "\r\n"
  print(x)
def display(entities):
  x_size = 0
  y_size = 0
  for i in entities:
    x = entities[i]
    if x.x > x_size:
      x_size = x.x
    if x.y > y_size:
      y_size = x.y
  arr = [ [" "] * x_size for i in range(y_size) ]
  for i in entities:
    x = entities[i]
    arr[x.y - 1][x.x - 1] = x.style + x.char + style.RESET
  print_map(arr)
def run_loop(calculate, initial_state):
  fd = sys.stdin.fileno()
  old_settings = termios.tcgetattr(fd)
  tty.setraw(sys.stdin)
  game_state = initial_state
  with raw(sys.stdin):
    with nonblocking(sys.stdin):
      try:
        while True:
            try:
                c = sys.stdin.read(4096)
                game_state = calculate(game_state, c)
                for i in game_state["entities"]:
                  game_state["entities"][i].update()
                display(game_state["entities"])
                if game_state["running"] == False:
                  break
            except IOError:
                print('not ready')
          
            time.sleep(.1)
      finally:
        termios.tcsetattr(fd, termios.TCSADRAIN, old_settings)
  