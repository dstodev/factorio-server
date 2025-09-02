'''Command to put a game server instance on the shelf.'''

from manage import game
from manage.util import dir_has_files


class Shelf:
    '''Put a game server instance on the shelf.'''

    def __init__(self, name: str) -> None:
        '''Initialize the Shelf command.

        :param name: The game to put on the shelf.
        :type name: str
        '''
        self.name = name

    def execute(self) -> None:
        '''Move the game server instance to the shelf.'''
        server_dir = game.server_dir(self.name)

        if dir_has_files(server_dir):
            shelf_dir = game.shelf_dir(self.name)

            shelf_dir.mkdir(parents=True, exist_ok=True)

            last_index = max(
                (int(d.stem) for d in shelf_dir.iterdir() if d.is_dir()),
                default=0)

            this_index = last_index + 1

            # server dir looks like: <repo>/server-files/<game-name>/hot
            # game.server_dir() returns the hot path. The hot path is created by
            # the server user in the Download command, meaning the host user may
            # not have permission to move it directly. Instead, move the parent
            # <game-name> directory to the shelf.
            server_dir.parent.rename(shelf_dir / str(this_index))
