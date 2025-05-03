'''Command to start a server.'''

from manage import game, paths
from manage.docker.container import Bind, GameContainer
from manage.docker.util import build_image


class Start:
    '''Start a server.'''

    def __init__(self, name: str) -> None:
        '''Initialize the Start command.

        :param name: The name of the server to start.
        :type name: str
        '''
        self.name = name

        # server_dir looks like: <repo>/server-files/<game-name>/hot
        self.server_dir = game.server_dir(name)

        binds = [
            Bind(host=self.server_dir.parent, guest='/game', writeable=True),
            Bind(host=game.download_script(name), guest='/download.sh', writeable=False),
        ]

        image, _logs = game.docker_image(name)

        self.container = GameContainer(name, image, binds)

    def execute(self) -> None:
        pass
