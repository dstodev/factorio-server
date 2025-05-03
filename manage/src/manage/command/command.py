'''Command interface.'''

from typing import Protocol, runtime_checkable


@runtime_checkable
class Command(Protocol):
    '''Interface for all actions.'''

    def execute(self) -> None:
        '''Perform an action.'''
