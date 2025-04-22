'''Defines the Schedule class to represent a sequence of actions to perform.'''

from manage.command.command import Command


class Schedule:
    '''Schedule represents a sequence of actions to perform.
    It is itself a Command, but cannot be added to itself.
    '''

    def __init__(self):
        '''Initialize a schedule with no actions.'''
        self._actions: list[Command] = []

    def add_action(self, action: Command) -> None:
        '''Add an action to the schedule.

        :param action: The action to add.
        :type action: Command
        :raises AssertionError: If action does not implement Command interface.
        '''
        assert action is not self, 'Cannot add self to schedule'
        assert isinstance(action, Command), 'Action must implement Command interface'
        self._actions.append(action)

    def execute(self):
        '''Perform all scheduled actions.'''
        for action in self._actions:
            action.execute()
