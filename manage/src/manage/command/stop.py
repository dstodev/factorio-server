'''Command to stop a game server.'''


from manage import game
from manage.command.command import Command
from manage.command.rcon import Rcon
from manage.command.schedule import Schedule
from manage.docker.container import GameContainer


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
        self.last_schedule = None

    def execute(self, add_cmds: list[Command] | None = None) -> None:
        '''Stop the game server.'''
        container = GameContainer(f'{self.name}-server', None)  # Only reattach

        cfg = game.cfg_data(self.name)

        try_save = Schedule()

        try:
            cfg_rcon = cfg['rcon']

            try:
                try_save.add_action(Rcon(self.name, [cfg_rcon['save-pre']]))
            except KeyError:
                pass

            try:
                try_save.add_action(Rcon(self.name, [cfg_rcon['save']]))
            except KeyError:
                pass

            try:
                try_save.add_action(Rcon(self.name, [cfg_rcon['stop']]))
            except KeyError:
                pass

        except KeyError:
            pass

        for cmd in add_cmds or []:
            try_save.add_action(cmd)

        try:
            # Try graceful shutdown first
            container.execute(['touch', '/tmp/stopfile'])
            try_save.execute()
            result = container.wait(timeout=self.timeout)
            self.last_result = result
            self.last_schedule = try_save

        except (RuntimeError, ConnectionError):
            container = GameContainer(f'{self.name}-server', None)
            if container.container is not None:
                container.container.stop()
                container.wait()
