'''Test Container actions.'''


from test.util_docker import clean_docker

import pytest
from docker.errors import BuildError, NotFound

import docker
from manage import PROJECT_NAME, paths
from manage.docker import build_image
from manage.docker.server_container import Bind, ServerContainer
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
        clean_docker(self.container_name)

    def test_container_build_image(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest',
                              'RUN ["echo", "Hello, World!"]')

        uncap(dockerfile)

        container = ServerContainer(self.container_name, dockerfile)

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

        container = ServerContainer(self.container_name, dockerfile)

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

        container = ServerContainer(self.container_name, dockerfile)

        assert container.name == self.container_name
        assert container.dockerfile == dockerfile

        with pytest.raises(BuildError):
            container.build_source_image()

    def test_can_base_on_rcon_image(self, tmp_file, uncap):
        expected_uid = 30120
        expected_gid = 30121
        expected_uname = 'test-user'
        expected_gname = 'test-group'

        rcon_image_name = 'test-can-base-on-rcon-image'

        dockerfile = tmp_file('test-image.dockerfile',
                              f'FROM {rcon_image_name}:latest',
                              f'RUN test "$(id -u)" = "{expected_uid}"',
                              f'RUN test "$(id -g)" = "{expected_gid}"',
                              f'RUN test "$(id -un)" = "{expected_uname}"',
                              f'RUN test "$(id -gn)" = "{expected_gname}"')

        uncap(dockerfile)

        build_args = {
            'user_id': expected_uid,
            'group_id': expected_gid,
            'user_name': expected_uname,
            'group_name': expected_gname,
        }

        try:
            build_image(paths.get('rcon') / 'Dockerfile', rcon_image_name, build_args)

            container = ServerContainer(self.container_name, dockerfile)

            result = container.build_source_image()
        finally:
            # clean_docker(rcon_image_name)
            pass

        uncap(result)

        assert f'FROM {rcon_image_name}:latest' in result
        assert 'Successfully built' in result
        assert f'Successfully tagged {self.container_name}:latest' in result

        assert f'{expected_uid}' in result
        assert f'{expected_gid}' in result
        assert expected_uname in result
        assert expected_gname in result

    def test_container_start(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        container = ServerContainer(self.container_name, dockerfile)

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

        guest_script = '/src/some-script.sh'

        bind = Bind(host=script, guest=guest_script, writeable=False)

        container = ServerContainer(self.container_name, dockerfile, binds=[bind])

        assert container.name == self.container_name
        assert container.dockerfile == dockerfile

        container.start(command=[guest_script])

        result = container.wait()

        assert isinstance(result, Result)
        assert result.exit_status == 1
        assert result.stdout == f'Path: {guest_script}\n'
        assert result.stderr == 'User: 0\n'

    def test_container_start_change_buildargs(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest',
                              'ARG user_id',
                              'ARG group_id',
                              'RUN addgroup -g $group_id testuser \\',
                              '  && adduser -u $user_id -D -G testuser testuser',
                              'USER testuser',)

        script = tmp_file('test-script.sh',
                          '#!/bin/sh',
                          'echo User: "$(id -u)"',
                          'echo Group: "$(id -g)"',
                          mode=0o755)

        uncap(dockerfile)
        uncap(script)

        guest_script = '/src/some-script.sh'

        bind = Bind(host=script, guest=guest_script, writeable=False)

        initial_uid = 30120
        initial_gid = 30121
        next_uid = 30122
        next_gid = 30123

        container = ServerContainer(self.container_name,
                                    dockerfile,
                                    build_args={'user_id': f'{initial_uid}',
                                                'group_id': f'{initial_gid}'},
                                    binds=[bind])

        assert container.name == self.container_name
        assert container.dockerfile == dockerfile

        container.start(command=[guest_script])

        result = container.wait()

        assert isinstance(result, Result)
        assert f'{initial_uid}' in result.stdout
        assert f'{initial_gid}' in result.stdout
        assert f'{next_uid}' not in result.stdout
        assert f'{next_gid}' not in result.stdout

        # Store the current ID because it is about to be replaced by a new image
        # with different build arguments. This is used to clean up the image
        # later.
        assert container.image is not None
        assert container.image.id is not None
        old_image_id = container.image.id

        container.build_args = {'user_id': f'{next_uid}', 'group_id': f'{next_gid}'}

        container.start(command=[guest_script])

        result = container.wait()

        assert isinstance(result, Result)
        assert f'{next_uid}' in result.stdout
        assert f'{next_gid}' in result.stdout
        assert f'{initial_uid}' not in result.stdout
        assert f'{initial_gid}' not in result.stdout

        clean_docker(old_image_id)

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

        guest_script = '/src/some-script.sh'

        bind = Bind(host=script, guest=guest_script, writeable=False)
        bind_writeable = Bind(host=writeable, guest='/writeable', writeable=True)
        bind_not_writeable = Bind(host=not_writeable, guest='/not-writeable', writeable=False)

        container = ServerContainer(self.container_name, dockerfile, binds=[bind,
                                                                            bind_writeable,
                                                                            bind_not_writeable])

        assert container.name == self.container_name
        assert container.dockerfile == dockerfile

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

        container = ServerContainer(self.container_name, dockerfile)

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

        container = ServerContainer(self.container_name, dockerfile)

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
        assert 'closing logfile writer' in log_content

    def test_container_reattach(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        log_file = dockerfile.parent / 'log.txt'

        uncap(dockerfile)
        uncap(log_file)

        container = ServerContainer(self.container_name, dockerfile)

        container.start(entrypoint=['echo', '1', '2'],
                        command=['3', '4'],
                        log_file=log_file)

        container = ServerContainer(self.container_name, dockerfile)

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
        result = container.wait()

        assert container.container is None
        assert container.monitor is None

        assert isinstance(result, Result)
        assert result.exit_status == 0
        assert result.stdout == '1 2 3 4\n'
        assert result.stderr == ''

        assert log_file.read_text() == f'1 2 3 4\n({PROJECT_NAME}) closing logfile writer\n'

    def test_execute(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        container = ServerContainer(self.container_name, dockerfile)

        container.start(command=['tail', '-f', '/dev/null'])  # Run until manually stopped

        assert container.container is not None

        result = container.execute(['pkill', 'tail'])  # SIGTERM the tail process

        assert result is not None
        assert result.exit_code == 0
        assert result.output == ''

        result = container.wait()

        assert container.container is None

        assert isinstance(result, Result)
        assert result.exit_status == 143  # 143 = 128 + 15 (SIGTERM)

    def test_execute_not_started(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        container = ServerContainer(self.container_name, dockerfile)

        with pytest.raises(RuntimeError, match='Container does not exist'):
            container.execute(['echo', 'Hello!'])

    def test_execute_not_running(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        container = ServerContainer(self.container_name, dockerfile)

        container.start(command=['echo', 'Hello!'])
        assert container.container is not None
        container.container.wait()

        with pytest.raises(RuntimeError, match='Container is not running'):
            container.execute(['echo', 'Hello!'])

    def test_wait_container_externally_removed(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest')

        uncap(dockerfile)

        container = ServerContainer(self.container_name, dockerfile)

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

        container = ServerContainer(self.container_name, dockerfile)

        assert container.wait() is None
