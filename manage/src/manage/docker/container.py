'''Represent a Docker container.'''

from pathlib import Path
from typing import NamedTuple


class Bind(NamedTuple):
    '''Represent a shared file or directory between host and container.

    Each Bind is similar to a `-v` parameter to e.g. `docker run -v host/path:guest/path`.
    '''
    host: Path
    guest: Path
    writeable: bool = False


class Container:
    '''Create and manage a Docker container.'''

    def __init__(self, name: str, image: Path, *binds: Bind) -> None:
        '''Initialize with persistent settings like name and image.'''
        self.name = name
        self.image = image
        self.binds = binds

        self.id = ''
