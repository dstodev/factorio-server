'''Represent a Docker container.'''

import os
from pathlib import Path
from typing import NamedTuple

from docker.errors import BuildError, NotFound
from docker.models.containers import Container
from docker.models.images import Image
from docker.types import Mount

import docker  # https://docker-py.readthedocs.io/en/stable/index.html
from manage.shell import Result


class Bind(NamedTuple):
    '''Represent a shared file or directory between host and container.

    Each Bind is similar to a `-v` parameter to e.g. `docker run -v host/path:guest/path`.
    '''
    host: Path
    guest: Path
    writeable: bool = False


class RunContainer:
    '''Create and manage a Docker container.'''

    def __init__(self, name: str, dockerfile_path: Path, binds: list[Bind] | None = None):
        '''Initialize with persistent settings like name and image.'''
        self.name = name
        self.dockerfile = dockerfile_path
        self.binds = binds or []

        self.image: Image | None = None

    def run(self,
            cmd: list[str] | None = None,
            entrypoint: list[str] | None = None,
            log_file: Path | None = None,
            wait: bool = True) -> Result | Container:
        '''Run the container with the given command.

        If wait is True, wait for the container to finish and return the result.
        Otherwise, return the container.
        '''
        self.build_source_image()
        assert self.image is not None

        mounts: list[Mount] = []

        for bind in self.binds:
            mounts.append(Mount(target=str(bind.guest),
                                source=str(bind.host),
                                type='bind',
                                read_only=not bind.writeable))

        client = docker.from_env()
        container = client.containers.run(self.image,
                                          command=cmd,
                                          entrypoint=entrypoint,
                                          detach=True,
                                          mounts=mounts,
                                          auto_remove=not wait)

        if log_file is not None:
            assert container.id is not None
            details = client.api.inspect_container(container.id)
            log_source = Path(details['LogPath'])
            os.link(log_source, log_file)  # TODO: This doesn't work because logs are owned by root

        if wait:
            wait_result = container.wait(timeout=10)

            assert 'StatusCode' in wait_result

            exit_status = wait_result['StatusCode']
            stdout = container.logs(stdout=True, stderr=False).decode('utf-8')
            stderr = container.logs(stdout=False, stderr=True).decode('utf-8')

            container.remove(force=True)

            return Result(exit_status, stdout, stderr)

        return container

    def build_source_image(self) -> str:
        '''Build the image, returning the output of the build process.
        Build errors are raised as exceptions.
        '''
        client = docker.from_env()

        try:
            self.image, logs = client.images.build(path=str(self.dockerfile.parent),
                                                   dockerfile=self.dockerfile.name,
                                                   tag=self.name,
                                                   rm=True)
        except BuildError as e:
            raise BuildError(e.msg, e.build_log) from None

        output = []

        for entry in logs:
            if isinstance(entry, dict):
                if 'stream' in entry:
                    output.append(entry['stream'])

        return ''.join(output)
