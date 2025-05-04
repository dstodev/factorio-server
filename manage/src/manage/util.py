'''Miscellaneous utility functions.'''

import datetime
from pathlib import Path

from manage.docker.container import GameContainer


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
