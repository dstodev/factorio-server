'''Command to put a game server instance on the shelf.'''

from manage import game


class Shelf:
    '''Put a game server instance on the shelf.'''

    def __init__(self, name: str) -> None:
        '''Initialize the shelf command.

        :param name: The game to put on the shelf.
        :type name: str
        '''
        self.name = name

    def execute(self) -> None:
        '''Move the game server instance to the shelf.'''

        server_dir = game.server_dir(self.name).parent
        shelf_dir = game.shelf_dir(self.name)

        shelf_dir.mkdir(parents=True, exist_ok=True)

        last_index = max(
            (int(d.stem) for d in shelf_dir.iterdir() if d.is_dir()),
            default=0)

        this_index = last_index + 1

        server_dir.rename(shelf_dir / str(this_index))
