'''Represent a Docker container.'''

import os
import tempfile
from multiprocessing import Process
from pathlib import Path
from typing import NamedTuple

from docker.errors import APIError, NotFound
from docker.models.containers import Container
from docker.models.images import Image
from docker.types import Mount

import docker  # https://docker-py.readthedocs.io/en/stable/index.html
from manage import PROJECT_NAME, game
from manage.docker.util import build_image
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


class GameContainer:
    '''Create and manage a Docker container for a server.'''

    def __init__(self,
                 name: str,
                 image: Image | None,
                 binds: list[Bind] | None = None):
        '''Initialize with persistent settings like name and image.


        :param name: The name of the container.
        :type name: str

        :param build_args: Build arguments to pass to the Dockerfile. Default is
            None.
        :type build_args: dict[str, str] | None

        :param binds: A list of Bind tuples defining files and directories to
            mount into the container.
        :type binds: list[Bind] | None
        '''
        self.name = name
        self.image: Image | None = image
        self.binds = binds or []

        self.container: Container | None = None
        try:
            client = docker.from_env()
            self.container = client.containers.get(name)
        except NotFound:
            pass

        self.monitor: Process | None = None
        self.log_file: Path | None = None

    def start(self,
              command: list[str] | None = None,
              entrypoint: list[str] | None = None,
              log_file: Path | None = None,
              auto_rm: bool = False):
        '''Start the container with the given command.

        A monitor process is spawned to watch the container's output and log it
        to a file. Even when no log_file is specified, the monitor lives until
        the container stops logging output.

        If auto_rm is True, after the monitor stops logging output, it will
        remove the container before exiting. This means the container is removed
        "at any time", and functions like self.container.logs() are no longer
        reliable, since logs will be removed alongside the container. (But the
        monitor will have already logged the output to a file.)

        :param command: The command to run in the container. Default is None.
        :type command: list[str] | None

        :param entrypoint: The entrypoint to run in the container. Default is
            None.
        :type entrypoint: list[str] | None

        :param log_file: The file to log the container's output to. Default is
            None.
        :type log_file: Path | None

        :param auto_rm: Whether to remove the container after it stops. Default
            is False.
        :type auto_rm: bool

        :raises RuntimeError: The container is already running. Use execute() to
            send additional commands.
        '''
        if self.container is not None:
            raise RuntimeError('Container is already running!')

        if self.image is None:
            raise RuntimeError('Container was not initialized with an image!')

        client = docker.from_env()

        self.container = client.containers.run(name=self.name,
                                               image=self.image,
                                               command=command,
                                               entrypoint=entrypoint,
                                               init=True,
                                               detach=True,
                                               mounts=self.build_mounts())
        self.log_file = log_file

        if log_file is None:
            # No need to store /dev/null to self.log_file, skipping self here
            log_file = Path(os.devnull)

        self.monitor = Process(target=monitor,
                               args=(self.container.id, log_file, auto_rm),
                               daemon=False)
        self.monitor.start()

    def build_mounts(self) -> list[Mount]:
        '''Return a list of mounts for the container based on self.binds'''
        binds = self.binds

        mounts: list[Mount] = []

        for bind in binds:
            mounts.append(Mount(source=str(bind.host),
                                target=str(bind.guest),
                                type='bind',
                                read_only=not bind.writeable))

        return mounts

    def execute(self, command: list[str] | None = None):
        '''Execute a command in the container.

        :param command: The command to run in the container. Default is None.
        '''
        if self.container is None:
            raise RuntimeError('Container does not exist!')

        if command is not None:
            try:
                result = self.container.exec_run(command, stdout=True, stderr=True)
                result = type(result)(result.exit_code, result.output.decode('utf-8'))
                return result
            except APIError as e:
                raise RuntimeError('Container is not running!') from e

    def wait(self, timeout: int = 10) -> Result | None:
        '''Wait for the container to finish.


        :param timeout: Time to wait before raising an exception. Default is 10.
        :type timeout: int
        :return: The exit status and logs from the container, or None if the
            container is not running.
        :rtype: Result | None
        '''
        container = self.container
        self.container = None

        if container is None:
            return None

        try:
            wait_result = container.wait(timeout=timeout)

            self.stop_monitor()

            assert 'StatusCode' in wait_result

            exit_status = wait_result['StatusCode']
            stdout = container.logs(stdout=True, stderr=False).decode('utf-8')
            stderr = container.logs(stdout=False, stderr=True).decode('utf-8')

            container.remove(force=True)

            return Result(exit_status, stdout, stderr)

        except NotFound:
            return None

    def stop_monitor(self, timeout: int = 10):
        '''Stop the monitor process.'''
        monitor_ = self.monitor
        self.monitor = None

        if monitor_:
            monitor_.join(timeout=timeout)
            assert monitor_.exitcode is not None, 'Monitor process is still running!'
