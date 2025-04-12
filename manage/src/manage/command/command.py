'''Command interface and Schedule class to execute commands.'''

from typing import Protocol, runtime_checkable


@runtime_checkable
class Command(Protocol):
    '''Interface for all commands.'''

    def execute(self) -> None:
        '''Execute the command.'''


class Schedule:
    '''A sequence of actions to perform.'''

    def __init__(self):
        '''Initialize a schedule with no actions.'''
        self._actions: list[Command] = []

    def add_action(self, action: Command) -> None:
        '''Add an action to the schedule.

        :param action: The action to add.
        :type action: Command
        :raises TypeError: If action does not implement Command interface.
        '''
        assert isinstance(action, Command), 'Action must implement Command interface'
        self._actions.append(action)

    def execute(self):
        '''Perform all scheduled actions.'''
        for action in self._actions:
            action.execute()
