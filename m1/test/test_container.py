'''Test Container actions.'''


import pytest
from docker.errors import NotFound
from manage import PROJECT_NAME_SHORT
from manage.docker import build_image as _build_image
from manage.docker.container import Bind, GameContainer
from manage.shell import Result
from manage.util import clean_docker

import docker


# Patch build_image to disable caching for all tests in this module
def build_image(*args, **kwargs):
    kwargs.pop('cache', None)
    return _build_image(*args, **kwargs, cache=False)


class TestContainer:
    def setup_method(self, method):
        '''Set up the container to a different value for each test so they do
        not interact with each other.
        '''
        # pylint: disable=attribute-defined-outside-init
        self.container_name = f'{method.__name__}'

    def teardown_method(self, _method):
        '''Clean up containers and images that were left behind.'''
        clean_docker(self.container_name)

    def test_container_start(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        image, _log = build_image(dockerfile, self.container_name)

        container = GameContainer(self.container_name, image)

        container.start(entrypoint=['echo', '1', '2'],
                        command=['3', '4'])

        assert container.container is not None
        assert container.container.name == self.container_name

        assert container.monitor is not None

        result = container.wait()

        assert container.container is None
        assert container.monitor is None

        assert isinstance(result, Result)
        assert result.exit_status == 0
        assert result.stdout == '1 2 3 4\n'
        assert result.stderr == ''

    def test_container_start_no_image(self):
        container = GameContainer(self.container_name, None)

        with pytest.raises(RuntimeError, match='Container was not initialized with an image'):
            container.start(entrypoint=['echo', '1', '2'],
                            command=['3', '4'])

    def test_container_start_bind_script(self, tmp_file, uncap):
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

        image, _log = build_image(dockerfile, self.container_name)

        guest_script = '/src/some-script.sh'

        bind = Bind(host=script, guest=guest_script, writeable=False)

        container = GameContainer(self.container_name, image, binds=[bind])

        container.start(command=[guest_script])

        result = container.wait()

        assert isinstance(result, Result)
        assert result.exit_status == 1
        assert result.stdout == f'Path: {guest_script}\n'
        assert result.stderr == 'User: 0\n'

    def test_container_start_bind_script_not_writable(self, tmp_file, uncap):
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

        image, _log = build_image(dockerfile, self.container_name)

        guest_script = '/src/some-script.sh'

        bind = Bind(host=script, guest=guest_script, writeable=False)
        bind_writeable = Bind(host=writeable, guest='/writeable', writeable=True)
        bind_not_writeable = Bind(host=not_writeable, guest='/not-writeable', writeable=False)

        container = GameContainer(self.container_name, image, binds=[bind,
                                                                     bind_writeable,
                                                                     bind_not_writeable])

        container.start(command=[guest_script])

        result = container.wait()

        assert isinstance(result, Result)
        assert result.exit_status != 0
        assert result.stdout == ''
        assert '/writeable' not in result.stderr
        assert '/not-writeable' in result.stderr

    def test_container_start_already_running(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        image, _log = build_image(dockerfile, self.container_name)

        container = GameContainer(self.container_name, image)

        container.start(command=['tail', '-f', '/dev/null'])  # Run until manually stopped

        with pytest.raises(RuntimeError, match='Container is already running'):
            container.start(command=['echo', 'Hello!'])

        assert container.container is not None
        container.container.stop(timeout=0)

    def test_container_log(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        log_file = dockerfile.parent / 'path/to/test.log'

        uncap(dockerfile)
        uncap(log_file)

        image, _log = build_image(dockerfile, self.container_name)

        container = GameContainer(self.container_name, image)

        assert container.log_file is None

        container.start(entrypoint=['/bin/sh', '-c'],
                        command=['echo Hello, && echo World! >&2'],
                        log_file=log_file)

        assert container.log_file == log_file

        result = container.wait()

        assert isinstance(result, Result)
        assert result.exit_status == 0
        assert result.stdout == 'Hello,\n'
        assert result.stderr == 'World!\n'

        with log_file.open() as f:
            log_content = f.read()

        assert 'Hello,' in log_content
        assert 'World!' in log_content
        assert 'closing logfile writer\n' in log_content

    def test_container_reattach(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        log_file = dockerfile.parent / 'log.txt'

        uncap(dockerfile)
        uncap(log_file)

        image, _log = build_image(dockerfile, self.container_name)

        container = GameContainer(self.container_name, image)
        monitor = container.monitor

        container.start(entrypoint=['echo', '1', '2'],
                        command=['3', '4'],
                        log_file=log_file)

        container = GameContainer(self.container_name, None)  # Reattach does not need image
        assert container.image is None

        assert container.container is not None
        assert container.container.name == self.container_name

        # Reattach behavior is intended for use between processes.
        # There is not a clean way to preserve the log file and monitor as
        # data between process runs (container reattachments), but the monitor
        # is still running & log file is still populating from the first run.
        assert container.monitor is None  # Cannot reattach monitor
        assert container.log_file is None  # Cannot reattach log file

        # Once the Docker container closes, the monitor stops and the log file
        # closes. It does not matter which process or instance of the
        # ServerContainer class stops the container.
        #
        # Hack: give the monitor to the new container so calling wait()
        # deterministically stops the monitor to guarantee logfile closure.
        container.monitor = monitor
        result = container.wait()

        assert container.container is None
        assert container.monitor is None

        assert isinstance(result, Result)
        assert result.exit_status == 0
        assert result.stdout == '1 2 3 4\n'
        assert result.stderr == ''

        assert log_file.read_text() == f'1 2 3 4\n({PROJECT_NAME_SHORT}) closing logfile writer\n'

    def test_execute(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest',
                              # Force unique cache for next layers;
                              # fixes frequent failures due to shared intermediate containers being deleted
                              f'RUN echo "{self.container_name}"',
                              'ARG user_id',
                              'ARG user_name',
                              'ARG group_id',
                              'ARG group_name',
                              'RUN addgroup -g $group_id $group_name \\',
                              '  && adduser -u $user_id -D -G $group_name $user_name',
                              'USER $user_name')

        uncap(dockerfile)

        expected_uid = 30120
        expected_gid = 30121

        build_args = {
            'user_id': expected_uid,
            'user_name': 'server-user',
            'group_id': expected_gid,
            'group_name': 'server-group'
        }

        image, _log = build_image(dockerfile, self.container_name, build_args)

        container = GameContainer(self.container_name, image)

        container.start(command=['tail', '-f', '/dev/null'])  # Run until manually stopped

        assert container.container is not None

        result = container.execute(['/bin/sh', '-c', 'echo "$(id -u):$(id -g)"'])

        assert result is not None
        assert result.exit_status == 0
        assert result.output == f'{expected_uid}:{expected_gid}\n'

        result = container.execute(['pkill', 'tail'])  # SIGTERM the tail process

        assert result is not None
        assert result.exit_status == 0
        assert result.output == ''

        result = container.wait()

        assert container.container is None

        assert isinstance(result, Result)
        assert result.exit_status == 143  # 143 = 128 + 15 (SIGTERM)

    def test_execute_not_started(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        image, _log = build_image(dockerfile, self.container_name)

        container = GameContainer(self.container_name, image)

        with pytest.raises(RuntimeError, match='Container does not exist'):
            container.execute(['echo', 'Hello!'])

    def test_execute_not_running(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        image, _log = build_image(dockerfile, self.container_name)

        container = GameContainer(self.container_name, image)

        container.start(command=['echo', 'Hello!'])
        assert container.container is not None
        container.container.wait()

        with pytest.raises(RuntimeError, match='Container is not running'):
            container.execute(['echo', 'Hello!'])

    def test_wait_container_externally_removed(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        image, _log = build_image(dockerfile, self.container_name)

        container = GameContainer(self.container_name, image)

        container.start()

        assert container.container is not None
        container_id = container.container.id

        client = docker.from_env()

        assert container_id is not None
        assert client.containers.get(container_id) is not None

        container.container.remove(force=True)

        with pytest.raises(NotFound):
            client.containers.get(container_id)

        assert container.wait() is None

    def test_wait_no_container(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        image, _log = build_image(dockerfile, self.container_name)

        container = GameContainer(self.container_name, image)

        assert container.wait() is None
