'''Represent a Docker container.'''

from pathlib import Path
from typing import NamedTuple

from docker.errors import BuildError
from docker.models.containers import Container
from docker.models.images import Image
from docker.types import Mount

import docker  # https://docker-py.readthedocs.io/en/stable/index.html
from manage.shell import Result

THIS_FILE = Path(__file__)
THIS_DIR = THIS_FILE.parent
LOG_ENTRYPOINT = THIS_DIR / 'log-entrypoint.sh'

assert LOG_ENTRYPOINT.is_file(), f'Log entrypoint script not found: {LOG_ENTRYPOINT}'


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

        binds = self.binds

        if log_file is not None:
            if entrypoint is None:
                entrypoint = []

            log_file.parent.mkdir(parents=True, exist_ok=True)
            log_file.touch(exist_ok=True, mode=0o644)
            assert log_file.is_file(), f'Log file not found: {log_file}'

            binds.append(Bind(host=log_file,
                              guest=Path('/log'),
                              writeable=True))

            binds.append(Bind(host=LOG_ENTRYPOINT,
                              guest=Path(f'/{LOG_ENTRYPOINT.name}'),
                              writeable=False))

            new_entrypoint = [f'/{LOG_ENTRYPOINT.name}', '/log', *entrypoint]
            entrypoint = new_entrypoint

        mounts: list[Mount] = []

        for bind in binds:
            mounts.append(Mount(source=str(bind.host),
                                target=str(bind.guest),
                                type='bind',
                                read_only=not bind.writeable))

        client = docker.from_env()
        container = client.containers.run(self.image,
                                          command=cmd,
                                          entrypoint=entrypoint,
                                          detach=True,
                                          mounts=mounts,
                                          auto_remove=not wait)

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
