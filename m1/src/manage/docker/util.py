'''Helpers for Docker image management.'''


from pathlib import Path
from typing import Iterable

from docker.models.containers import Container
from docker.models.images import Image
from manage.shell import Result
from python_on_whales import docker as whale

import docker


def build_image(dockerfile: Path,
                name: str,
                build_args: dict[str, str] | None = None,
                verbose: bool = False,
                cache: bool = True) -> tuple[Image, str]:
    '''Build an image.'''
    build_args = {k: str(v) for k, v in (build_args or {}).items()}
    # use python-on-whales build with cli to support buildkit
    client = docker.from_env()

    log_stream = whale.build(context_path=str(dockerfile.parent.resolve()),
                             file=dockerfile.resolve(),
                             tags=[name],
                             build_args=build_args,
                             stream_logs=True,
                             cache=cache)

    assert isinstance(log_stream, Iterable)

    log_full = ''

    for line in log_stream:
        line = line.strip()
        if line:
            log_full += f'{line}\n'
            if verbose:
                print(f':: {line}')

    image = client.images.get(name)
    assert isinstance(image, Image)

    return image, log_full


def wait_for_container(container: Container, timeout: int = 10) -> Result:
    '''Wait for a container to exit, then return its exit status and logs.'''
    wait_result = container.wait(timeout=timeout)

    exit_status = wait_result['StatusCode']
    stdout = container.logs(stdout=True, stderr=False).decode('utf-8')
    stderr = container.logs(stdout=False, stderr=True).decode('utf-8')

    return Result(exit_status, stdout, stderr)
