'''Represent a Docker container.'''

import os
from multiprocessing import Process
from pathlib import Path
from typing import NamedTuple

from docker.errors import BuildError
from docker.models.containers import Container
from docker.models.images import Image
from docker.types import Mount

import docker  # https://docker-py.readthedocs.io/en/stable/index.html
from manage import PROJECT_NAME
from manage.shell import Result


def monitor(container_id: str, log_path: Path, auto_rm: bool = False):  # pragma: no cover
    '''Monitor a container, writing stdout & stderr to a file, and optionally
    clean up the container when done.

    This intended to run in a separate process to continue running as long as
    the container is running.
    '''
    client = docker.from_env()
    container = client.containers.get(container_id)

    generator = container.logs(stream=True, follow=True, stdout=True, stderr=True)

    log_path.parent.mkdir(parents=True, exist_ok=True)

    with open(log_path, 'a', encoding='utf-8') as logfile:
        for line in generator:
            logfile.write(f'{line.decode("utf-8")}')
            logfile.flush()

        if auto_rm:
            logfile.write(f'({PROJECT_NAME}) removing container {container_id}\n')
            container.remove()

        logfile.write(f'({PROJECT_NAME}) closing logfile writer\n')


class Bind(NamedTuple):
    '''Represent a shared file or directory between host and container.

    Each Bind is similar to a `-v` parameter to e.g. `docker run -v host/path:guest/path`.
    '''
    host: Path
    guest: Path | str
    writeable: bool = False


class RunContainer:
    '''Create and manage a Docker container.'''

    def __init__(self,
                 name: str,
                 dockerfile_path: Path,
                 build_args: dict[str, str] | None = None,
                 binds: list[Bind] | None = None):
        '''Initialize with persistent settings like name and image.'''
        self.name = name
        self.dockerfile = dockerfile_path
        self.build_args = build_args or {}
        self.binds = binds or []

        self.image: Image | None = None

        self.monitor: Process | None = None

    def run(self,
            command: list[str] | None = None,
            entrypoint: list[str] | None = None,
            log_file: Path | None = None,
            wait: bool = True) -> Result | Container:
        '''Run the container with the given command.

        If wait is True, wait for the container to finish and return the result.
        Otherwise, return the container.
        '''
        _build_output = self.build_source_image()
        assert self.image is not None

        binds = self.binds

        mounts: list[Mount] = []

        for bind in binds:
            mounts.append(Mount(source=str(bind.host),
                                target=str(bind.guest),
                                type='bind',
                                read_only=not bind.writeable))

        client = docker.from_env()
        container = client.containers.run(self.image,
                                          command=command,
                                          entrypoint=entrypoint,
                                          detach=True,
                                          mounts=mounts)

        if log_file is None:
            log_file = Path(os.devnull)

        self.monitor = Process(target=monitor,
                               args=(container.id, log_file, not wait),
                               daemon=False)
        self.monitor.start()

        if wait:
            wait_result = container.wait(timeout=10)
            self.monitor.join(timeout=10)
            assert self.monitor.exitcode is not None
            self.monitor = None

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
                                                   buildargs=self.build_args,
                                                   rm=True)
        except BuildError as e:
            raise BuildError(e.msg, e.build_log) from None

        output = []

        for entry in logs:
            if isinstance(entry, dict):
                if 'stream' in entry:
                    output.append(entry['stream'])

        return ''.join(output)
