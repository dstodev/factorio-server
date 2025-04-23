'''Test Container actions.'''


from pathlib import Path

import pytest
from docker.errors import BuildError

from manage.docker.container import Bind, Container


def test_container(tmp_file, uncap):
    '''Test the container class.'''
    image = tmp_file('test_image.dockerfile',
                     'FROM alpine:latest',
                     'CMD ["echo", "Hello, World!"]')

    container = Container('test_container', image)

    uncap(image)

    assert container.name == 'test_container'
    assert container.image_path == image

    result = container.build_image()

    uncap(result)

    assert 'Successfully built' in result
    assert 'Successfully tagged test_container:latest' in result


def test_container_error_invalid_image(tmp_file, uncap):
    '''Test the container class with an invalid image.'''
    image = tmp_file('test_image.dockerfile',
                     'FROM invalid_image:latest',
                     'CMD ["echo", "Hello, World!"]')

    container = Container('test_container', image)

    uncap(image)

    assert container.name == 'test_container'
    assert container.image_path == image

    with pytest.raises(BuildError) as e:
        container.build_image()
