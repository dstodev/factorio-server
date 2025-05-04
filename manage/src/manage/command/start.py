'''Command to start a server.'''

from manage import game
from manage.docker.container import Bind, GameContainer


class Start:
    '''Start a server by running its start script in a persistent container.'''

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
            Bind(host=game.start_script(name), guest='/start.sh', writeable=False),
        ]

        image, _logs = game.docker_image(name)

        self.container = GameContainer(f'{name}-server', image, binds)

    def execute(self, auto_rm: bool = True) -> None:
        '''Start the server.

        auto_rm is not normally accessible since function implements the Command
        protocol, which does not take any arguments. The flag exists for tests.
        '''
        parent_dir = self.server_dir.parent
        parent_dir.mkdir(parents=True, exist_ok=True)

        # /game aligns with guest value for server_dir.parent bind mount in __init__
        guest_server_dir = '/game/hot'

        rcon = game.rcon_password(self.name, new=True)

        command = ' && '.join([
            f'mkdir -p {guest_server_dir}',
            'cp /start.sh /tmp/start.sh',
            f'while [ ! -f /tmp/stopfile ]; do /tmp/start.sh {guest_server_dir} {rcon}; done',
            'rm -f /tmp/stopfile',
        ])

        parent_dir = self.server_dir.parent
        parent_dir.mkdir(parents=True, exist_ok=True)

        self.container.start(entrypoint=['/bin/sh', '-c'],
                             command=[command],
                             log_file=game.logs_dir(self.name) / 'server.log',
                             auto_rm=auto_rm)
