'''Docker utility functions for testing.'''

from manage.docker import GameContainer


class PKillTail:
    '''Run 'pkill tail' as a Command.'''

    def __init__(self, container: GameContainer):
        self.container = container

    def execute(self):
        self.container.execute(['pkill', 'tail'])
