'''Helpers for Docker image management.'''


from pathlib import Path

from docker.models.images import Image

import docker


def build(dockerfile: Path,
          name: str,
          build_args: dict[str, str] | None = None) -> tuple[Image, str]:
    '''Build an image.'''
    client = docker.from_env()
    image, logs = client.images.build(path=str(dockerfile.parent),
                                      dockerfile=dockerfile.name,
                                      tag=name,
                                      buildargs=build_args,
                                      rm=True)

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
