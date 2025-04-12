'''Test the Command interface & related Schedule class.'''

from manage.command import Command, Schedule


class NotCommand:
    '''Does not implement the Command interface.'''


class DummyCommand:
    '''Implements the Command interface.'''

    def execute(self):
        '''No-op'''


class StoreExecuted:
    '''Store how many times this command is executed.'''

    def __init__(self):
        '''Start at zero.'''
        self._count = 0

    def execute(self):
        '''Increment the call count.'''
        self._count += 1

    def count(self):
        '''Get call count.'''
        return self._count


def test_command():
    assert not isinstance(NotCommand(), Command)
    assert isinstance(DummyCommand(), Command)


def test_schedule():
    schedule = Schedule()

    cmd1 = StoreExecuted()
    schedule.add_action(cmd1)
    schedule.execute()
    assert cmd1.count() == 1

    cmd2 = StoreExecuted()
    schedule.add_action(cmd2)
    schedule.execute()
    assert cmd1.count() == 2
    assert cmd2.count() == 1
