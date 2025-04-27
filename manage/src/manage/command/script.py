'''Command to run a script.'''

import pathlib

from manage import shell


class ScriptCommand:
    '''Implements the Command interface to run a script.'''

    def __init__(self, script: pathlib.Path, *args: str) -> None:
        '''Initialize the script command.

        :param script: The script to run when executed.
        :type script: pathlib.Path
        :param args: Arguments to pass to the script.
        :type args: str
        '''
        self.script = script
        self.args = args
        self.last_result = None

    def execute(self) -> None:
        '''Run the script with stored arguments.'''
        self.last_result = shell.run_file(self.script, *self.args)

    def get_result(self) -> shell.Result | None:
        '''Get the result of the last call to execute().

        :return: The result of the last call to execute(), if performed.
        :rtype: shell.Result
        '''
        return self.last_result
