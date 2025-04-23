'''Represent a Docker container.'''

from pathlib import Path
from typing import NamedTuple

from docker.errors import BuildError
from docker.models.images import Image

import docker


class Bind(NamedTuple):
    '''Represent a shared file or directory between host and container.

    Each Bind is similar to a `-v` parameter to e.g. `docker run -v host/path:guest/path`.
    '''
    host: Path
    guest: Path
    writeable: bool = False


class Container:
    '''Create and manage a Docker container.'''

    def __init__(self, name: str, image_path: Path, *binds: Bind):
        '''Initialize with persistent settings like name and image.'''
        self.name = name
        self.image_path = image_path
        self.binds = binds

        self.image: Image | None = None

    def run(self, *cmd: str, entrypoint: str = ''):
        '''Run the container with the given command.'''
        self.build_image()
        client = docker.from_env()

    def build_image(self) -> str:
        '''Build the image, returning the output of the build process.
        Build errors are raised as exceptions.
        '''
        client = docker.from_env()  # https://docker-py.readthedocs.io/en/stable/client.html

        try:
            self.image, logs = client.images.build(path=str(self.image_path.parent),
                                                   dockerfile=self.image_path.name,
                                                   tag=self.name,
                                                   rm=True)
        except BuildError as e:
            raise BuildError(e.msg, e.build_log) from None

        output = []

        for entry in logs:
            print(entry)  # TODO: Remove this line once confident all keys are covered

            if isinstance(entry, dict):
                if 'stream' in entry:
                    output.append(entry['stream'])

        return ''.join(output)
