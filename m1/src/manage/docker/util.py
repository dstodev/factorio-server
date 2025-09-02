'''Helpers for Docker image management.'''


from pathlib import Path

from docker.models.containers import Container
from docker.models.images import Image

import docker
from manage.shell import Result


def build_image(dockerfile: Path,
                name: str,
                build_args: dict[str, str] | None = None) -> tuple[Image, str]:
    '''Build an image.'''
    build_args = {k: str(v) for k, v in (build_args or {}).items()}
    client = docker.from_env()
    image, logs = client.images.build(path=str(dockerfile.parent),
                                      dockerfile=dockerfile.name,
                                      tag=name,
                                      buildargs=build_args,
                                      rm=True,
                                      forcerm=True)

    return image, build_logs_to_str(logs)


def build_logs_to_str(logs) -> str:
    '''Process the logs from a Docker image build:

    .. code-block:: python
        image, logs = client.images.build(...)
        log_str = build_logs_to_str(logs)
    '''
    output = []

    for entry in logs:
        if isinstance(entry, dict):
            if 'stream' in entry:
                output.append(entry['stream'])

    return ''.join(output)


def wait_for_container(container: Container, timeout: int = 10) -> Result:
    '''Wait for a container to exit, then return its exit status and logs.'''
    wait_result = container.wait(timeout=timeout)

    exit_status = wait_result['StatusCode']
    stdout = container.logs(stdout=True, stderr=False).decode('utf-8')
    stderr = container.logs(stdout=False, stderr=True).decode('utf-8')

    return Result(exit_status, stdout, stderr)
