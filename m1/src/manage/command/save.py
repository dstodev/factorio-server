'''Command to save a game server.'''

from manage.command.command import Command
from manage.command.rcon import Rcon
from manage.command.schedule import Schedule


class Save:
    '''Use RCON to save a game server.

    Some games, like Minecraft, use a sequence like:

    1. Disable write to disk
    2. Flush outstanding data to disk
    3. Perform backup
    4. Re-enable write to disk

    To support this, Save accepts a Command "between_cmd" to run between steps 2
    and 3.

    To use this, in the server.json entry for the save command, use the extended
    save command syntax:

    "rcon": {
        "save": "a,b;c"
    }

    RCON commands before the semicolon are run before the provided Command
    "between_cmd", and commands after the semicolon are run after it. Multiple
    commands are supported by separating them with commas.

    If no Command is provided, the RCON commands are run in sequence.

    If no semicolon is provided, the Command is run after the RCON commands.

    All RCON commands run with ignored exceptions, meaning the provided Command
    will always run.
    '''

    def __init__(self, name: str, between_cmd: Command) -> None:
        '''Initialize the save command.

        :param name: The game to save.
        :type name: str
        '''
        self.name = name

        self.last_result = None
        self.last_save = None
        self.between_cmd = between_cmd

# TODO: Finish this
