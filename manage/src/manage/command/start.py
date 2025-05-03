'''Command to start a server.'''

from manage import game, paths
from manage.docker.util import build_image


class Start:
    '''Start a server.'''

    def __init__(self, name: str) -> None:
        '''Initialize the Start command.

        :param name: The name of the server to start.
        :type name: str
        '''
        self.name = name

    def execute(self) -> None:
        # Build the base image (rcon:latest)
        build_image(paths.get('rcon') / 'Dockerfile', 'rcon', game.build_args(self.name))
