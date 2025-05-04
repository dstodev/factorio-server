'''Command to use the RCON client.'''

from manage.docker.container import GameContainer


class Rcon:
    '''Send an RCON command to a server.'''

    def __init__(self, name: str, command: list[str], password: str | None = None) -> None:
        '''Initialize the RCON command.

        :param name: The game to use.
        :type name: str
        :param command: The command to send.
        :type command: str
        '''
        self.name = name
        self.command = command
        self.last_result = None
        self.password = password

    def execute(self) -> None:
        '''Run the RCON command.

        :raises RuntimeError: The server is not running.
        '''
        container = GameContainer(self.name, None)  # Only reattach

        argstr = ' '.join(self.command)
        cmdstr = f'rcon {argstr}'

        password = self.password

        if password is not None:
            cmdstr = f'echo "{password}" | {cmdstr}'

        self.last_result = container.execute(['/bin/sh', '-c', cmdstr])
