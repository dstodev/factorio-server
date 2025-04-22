'''Command interface.'''

from typing import Protocol, runtime_checkable


@runtime_checkable
class Command(Protocol):
    '''Interface for all commands.'''

    def execute(self) -> None:
        '''Execute the command.'''
