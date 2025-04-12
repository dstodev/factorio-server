'''Pytest fixture to help print messages from tests even when output is captured.'''

import pytest


@pytest.fixture
def uncap(capsys, request) -> '_Printer':
    '''Print a message even if output is normally captured.

    Example:

    .. code-block:: python
        def test_something(uncap):
            uncap('Hello,')
            uncap('World!')

    This will print:
    .. code-block:: text
        test/test_module.py::test_something:
          Hello,
          World!
    '''
    return _Printer(request.node.nodeid, capsys)


class _Printer:
    '''Stateful functor instantiated once per test, but callable multiple times.'''

    def __init__(self, prefix: str, capsys):
        '''Initialize the printer by providing the capsys fixture from pytest.'''
        self._prefix = prefix
        self._capsys = capsys
        self._printed_preamble = False

    def __call__(self, message: str, **kwargs):
        '''Print a message.'''
        with self._capsys.disabled():
            self.preamble()
            print(f'  {message}', **kwargs)

    def preamble(self):
        '''Print the preamble once.'''
        if not self._printed_preamble:
            print(f'\n{self._prefix}:')
            self._printed_preamble = True
