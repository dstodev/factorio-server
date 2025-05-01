'''Docker utility functions for testing.'''

from docker.errors import NotFound

import docker


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
