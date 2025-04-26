'''Test Container actions.'''


from pathlib import Path

import pytest
from docker.errors import BuildError, NotFound

import docker
from manage.docker.run_container import Bind, RunContainer
from manage.shell import Result


class TestContainer:
    def setup_method(self, method):
        '''Set up the container to a different value for each test so they do
        not interact with each other.
        '''
        # pylint: disable=attribute-defined-outside-init
        self.container_name = f'{method.__name__}'

    def teardown_method(self, _method):
        '''Clean up containers and images that were left behind.'''
        client = docker.from_env()

        try:
            container = client.containers.get(self.container_name)  # Raises NotFound
            container.remove(force=True)  # pragma: no cover
        except NotFound:
            pass

        images = client.images.list(name=self.container_name, all=True)

        for image in images:
            image.remove(force=True)

    def test_container_build_image(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest',
                              'RUN ["echo", "Hello, World!"]')

        container = RunContainer(self.container_name, dockerfile)

        uncap(dockerfile)

        assert container.name == self.container_name
        assert container.dockerfile == dockerfile

        result = container.build_source_image()

        uncap(result)

        assert 'Successfully built' in result
        assert f'Successfully tagged {self.container_name}:latest' in result

    def test_container_rebuild_image(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest',
                              'RUN ["echo", "Hello,"]')

        container = RunContainer(self.container_name, dockerfile)

        uncap(dockerfile)

        assert container.name == self.container_name
        assert container.dockerfile == dockerfile

        result = container.build_source_image()

        uncap(result)

        assert 'Successfully built' in result
        assert 'Hello,' in result
        assert 'World!' not in result
        assert f'Successfully tagged {self.container_name}:latest' in result

        with dockerfile.open('a') as f:
            f.write('RUN ["echo", "World!"]\n')

        result = container.build_source_image()

        uncap(result)

        assert 'Successfully built' in result
        assert 'World!' in result
        assert f'Successfully tagged {self.container_name}:latest' in result

    def test_container_error_invalid_image(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM invalid_image:latest',
                              'RUN ["echo", "Hello, World!"]')

        container = RunContainer(self.container_name, dockerfile)

        uncap(dockerfile)

        assert container.name == self.container_name
        assert container.dockerfile == dockerfile

        with pytest.raises(BuildError):
            container.build_source_image()

    def test_container_run(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        container = RunContainer(self.container_name, dockerfile)

        uncap(dockerfile)

        result = container.run(entrypoint=['echo', '1', '2'], cmd=['3', '4'])

        assert isinstance(result, Result)
        assert result.exit_status == 0
        assert result.stdout == '1 2 3 4\n'
        assert result.stderr == ''

    def test_container_run_bind_script(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        script = tmp_file('test-script.sh',
                          '#!/bin/sh',
                          'echo "Path: $(readlink -f -- "$0")"',
                          'echo "User: $(id -u)" >&2',
                          'exit 1',
                          mode=0o744)

        uncap(dockerfile)
        uncap(script)

        guest_path = Path('/src/some-script.sh')

        bind = Bind(host=script, guest=guest_path, writeable=False)

        container = RunContainer(self.container_name, dockerfile, binds=[bind])

        assert container.name == self.container_name
        assert container.dockerfile == dockerfile

        result = container.run(cmd=[str(guest_path)])

        assert isinstance(result, Result)
        assert result.exit_status == 1
        assert result.stdout == f'Path: {guest_path}\n'
        assert result.stderr == 'User: 0\n'

    def test_container_log(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        log_file = dockerfile.parent / 'test.log'

        uncap(dockerfile)
        uncap(log_file)

        container = RunContainer(self.container_name, dockerfile)

        result = container.run(cmd=['echo', 'Hello, World!'], log_file=log_file)

        assert isinstance(result, Result)
        assert result.exit_status == 0
        assert result.stdout == 'Hello, World!\n'
        assert result.stderr == ''

        with log_file.open() as f:
            log_content = f.read()

        assert 'Hello, World!' in log_content
