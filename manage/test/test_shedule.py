'''Test the Command interface & related Schedule class.'''

import pytest

from manage.command import Command, Schedule


class NotCommand:
    '''Does not implement the Command interface.'''


class DummyCommand:
    '''Implements the Command interface.'''

    def execute(self):
        '''No-op'''


class CallCount:
    '''Implements the Command interface, storing the number of times execute() is called.'''

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
    cmd1 = CallCount()
    assert cmd1.count() == 0

    schedule = Schedule()
    schedule.add_action(cmd1)
    schedule.execute()
    assert cmd1.count() == 1

    cmd2 = CallCount()
    assert cmd2.count() == 0

    schedule.add_action(cmd2)
    schedule.execute()
    assert cmd1.count() == 2
    assert cmd2.count() == 1


def test_schedule_duplicate():
    cmd = CallCount()
    schedule = Schedule()
    schedule.add_action(cmd)
    schedule.add_action(cmd)
    schedule.execute()
    assert cmd.count() == 2


def test_schedule_type_error():
    schedule = Schedule()

    with pytest.raises(TypeError, match='Action must implement Command interface'):
        schedule.add_action(NotCommand())  # type: ignore


def test_schedules_are_commands():
    cmd = CallCount()
    schedule1 = Schedule()
    schedule1.add_action(cmd)

    schedule2 = Schedule()
    schedule2.add_action(schedule1)

    schedule2.execute()

    assert cmd.count() == 1

    schedule2.add_action(cmd)
    schedule2.execute()

    assert cmd.count() == 3


def test_add_self_as_action():
    schedule = Schedule()

    with pytest.raises(ValueError, match='Cannot add self to schedule'):
        schedule.add_action(schedule)
