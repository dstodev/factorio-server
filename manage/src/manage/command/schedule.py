'''Command to perform a sequence of actions.'''

from manage.command.command import Command


class Schedule:
    '''Schedule represents a sequence of actions to perform.
    Schedule is a Command, but you cannot add a schedule to itself.
    '''

    def __init__(self):
        '''Initialize a schedule with no actions.'''
        self._actions: list[Command] = []

    def add_action(self, action: Command) -> None:
        '''Add an action to the schedule.

        :param action: The action to add.
        :type action: Command
        :raises AssertionError: Action does not implement Command interface.
        '''
        if action is self:
            raise ValueError('Cannot add self to schedule!')
        if not isinstance(action, Command):
            raise TypeError('Action must implement Command interface!')
        self._actions.append(action)

    def execute(self):
        '''Perform all scheduled actions.'''
        for action in self._actions:
            action.execute()
