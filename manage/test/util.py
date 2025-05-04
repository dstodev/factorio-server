'''Docker utility functions for testing.'''

from docker.errors import NotFound

import docker
from manage.docker import GameContainer


def clean_docker(name: str):  # pragma: no cover
    '''Remove any containers or images with the given name.
    This is useful to clean up after tests that create Docker containers or images.
    '''
    client = docker.from_env()

    try:
        container = client.containers.get(name)
        container.remove(force=True)
    except NotFound:
        pass

    try:
        client.images.remove(name, force=True)
    except NotFound:
        pass


class PKillTail:
    '''Run 'pkill tail' as a Command.'''

    def __init__(self, container: GameContainer):
        self.container = container

    def execute(self):
        self.container.execute(['pkill', 'tail'])
