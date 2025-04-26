'''Test Container actions.'''


from pathlib import Path

import pytest
from docker.errors import BuildError, NotFound
from docker.models.containers import Container

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
            container.wait()  # pragma: no cover
        except NotFound:
            pass

        try:
            client.images.remove(self.container_name, force=True)
        except NotFound:
            pass

    def test_container_build_image(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest',
                              'RUN ["echo", "Hello, World!"]')

        uncap(dockerfile)

        container = RunContainer(self.container_name, dockerfile)

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

        uncap(dockerfile)

        container = RunContainer(self.container_name, dockerfile)

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

        uncap(dockerfile)

        container = RunContainer(self.container_name, dockerfile)

        assert container.name == self.container_name
        assert container.dockerfile == dockerfile

        with pytest.raises(BuildError):
            container.build_source_image()

    def test_container_run(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        container = RunContainer(self.container_name, dockerfile)

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

        guest_script = Path('/src/some-script.sh')

        bind = Bind(host=script, guest=guest_script, writeable=False)

        container = RunContainer(self.container_name, dockerfile, binds=[bind])

        assert container.name == self.container_name
        assert container.dockerfile == dockerfile

        result = container.run(cmd=[str(guest_script)])

        assert isinstance(result, Result)
        assert result.exit_status == 1
        assert result.stdout == f'Path: {guest_script}\n'
        assert result.stderr == 'User: 0\n'

    def test_container_run_bind_script_not_writable(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        script = tmp_file('test-script.sh',
                          '#!/bin/sh',
                          'touch /writeable/file',
                          'touch /not-writeable/file',
                          mode=0o744)

        writeable = dockerfile.parent / 'writeable'
        writeable.mkdir(parents=True, exist_ok=True)
        assert writeable.is_dir()

        not_writeable = dockerfile.parent / 'not-writeable'
        not_writeable.mkdir(parents=True, exist_ok=True)
        assert not_writeable.is_dir()

        uncap(dockerfile)
        uncap(script)
        uncap(writeable)
        uncap(not_writeable)

        guest_script = Path('/src/some-script.sh')

        bind = Bind(host=script, guest=guest_script, writeable=False)
        bind_writeable = Bind(host=writeable, guest=Path('/writeable'), writeable=True)
        bind_not_writeable = Bind(host=not_writeable, guest=Path('/not-writeable'), writeable=False)

        container = RunContainer(self.container_name, dockerfile, binds=[bind,
                                                                         bind_writeable,
                                                                         bind_not_writeable])

        assert container.name == self.container_name
        assert container.dockerfile == dockerfile

        result = container.run(cmd=[str(guest_script)])

        assert isinstance(result, Result)
        assert result.exit_status != 0
        assert result.stdout == ''
        assert '/writeable' not in result.stderr
        assert '/not-writeable' in result.stderr

    def test_container_log(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        log_file = dockerfile.parent / 'path/to/test.log'

        uncap(dockerfile)
        uncap(log_file)

        container = RunContainer(self.container_name, dockerfile)

        result = container.run(entrypoint=['/bin/sh', '-c'],
                               cmd=['echo Hello, && echo World! >&2'],
                               log_file=log_file)

        assert isinstance(result, Result)
        assert result.exit_status == 0
        assert result.stdout == 'Hello,\n'
        assert result.stderr == 'World!\n'

        with log_file.open() as f:
            log_content = f.read()

        assert 'Hello,' in log_content
        assert 'World!' in log_content
        assert 'closing logfile writer' in log_content

    def test_container_run_no_wait(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        container = RunContainer(self.container_name, dockerfile)

        result = container.run(entrypoint=['/bin/sh', '-c'],
                               cmd=['echo Hello, && echo World! >&2'],
                               wait=False)

        assert isinstance(result, Container)

        docker_container = result
        wait_result = docker_container.wait(timeout=10)
        assert container.monitor is not None
        container.monitor.join(timeout=10)
        assert container.monitor.exitcode is not None

        assert 'StatusCode' in wait_result

        exit_status = wait_result['StatusCode']

        assert exit_status == 0

        assert docker_container.id is not None

        # Assert container is deleted after exit
        with pytest.raises(NotFound):
            client = docker.from_env()
            client.containers.get(docker_container.id)

    def test_container_run_log_no_wait(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        log_file = dockerfile.parent / 'test.log'

        uncap(dockerfile)
        uncap(log_file)

        container = RunContainer(self.container_name, dockerfile)

        result = container.run(entrypoint=['/bin/sh', '-c'],
                               cmd=['echo Hello, && echo World! >&2'],
                               log_file=log_file,
                               wait=False)

        assert isinstance(result, Container)

        docker_container = result
        wait_result = docker_container.wait(timeout=10)
        assert container.monitor is not None
        container.monitor.join(timeout=10)
        assert container.monitor.exitcode is not None

        with log_file.open() as f:
            log_content = f.read()

        assert 'Hello,' in log_content
        assert 'World!' in log_content
        assert f'removing container {docker_container.id}' in log_content
        assert 'closing logfile writer' in log_content

        assert 'StatusCode' in wait_result

        exit_status = wait_result['StatusCode']

        assert exit_status == 0

        assert docker_container.id is not None

        # Assert container is deleted after exit
        with pytest.raises(NotFound):
            client = docker.from_env()
            client.containers.get(docker_container.id)
