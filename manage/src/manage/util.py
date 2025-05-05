'''Miscellaneous utility functions.'''

import datetime
from pathlib import Path

from docker.errors import NotFound

import docker


def timestamp() -> str:
    '''Get the current timestamp in UTC.

    :return: The current timestamp in UTC.
    :rtype: str
    '''
    return datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%d_%H-%M-%SZ')


def dir_has_files(path: Path) -> bool:
    '''Check if a directory has files.

    :param path: The directory to check.
    :type path: Path

    :return: True if the directory exists and is not empty, False otherwise.
    :rtype: bool
    '''
    return path.is_dir() and any(path.iterdir())


def clean_docker(name: str):  # pragma: no cover
    '''Remove any containers or images with the given name.
    This is useful to clean up after commands that create Docker containers or images.
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
