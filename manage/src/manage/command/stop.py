'''Command to stop a game server.'''


from requests.exceptions import ConnectionError as RequestsConnectionError

from manage import game
from manage.command.rcon import Rcon
from manage.command.schedule import Schedule
from manage.docker.container import GameContainer
from manage.rcon import RconError


class Stop:
    '''Stop a game server.

    Does nothing if the server is not running.
    '''

    def __init__(self, name: str, timeout_s: int = 10) -> None:
        '''Initialize the stop command.

        :param name: The game to stop.
        :type name: str
        '''
        self.name = name
        self.timeout = timeout_s
        self.last_result = None

        self.try_rcon = Schedule()

        try:
            cfg_rcon = game.cfg_data(name)['rcon']

            for key in ('save-pre', 'save', 'stop'):
                try:
                    self.try_rcon.add_action(Rcon(name, [cfg_rcon[key]]))
                except KeyError:
                    pass
        except KeyError:
            pass

    def execute(self) -> None:
        '''Stop the game server.'''
        container = GameContainer(f'{self.name}-server', None)  # Only reattach

        try:
            # Try graceful shutdown first
            container.execute(['touch', '/tmp/stopfile'])
            self.try_rcon.execute()
            result = container.wait(timeout=self.timeout)
            self.last_result = result

        except (RconError, RequestsConnectionError):
            container = GameContainer(f'{self.name}-server', None)
            if container.container is not None:
                container.container.stop()
                container.wait()
