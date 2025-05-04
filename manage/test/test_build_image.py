'''Test game.docker_image() behaviors.'''

from test.util_docker import clean_docker

import pytest
from docker.errors import BuildError

from manage.docker import build_image


class TestBuildImage:
    def setup_method(self, method):
        '''Set up the container to a different value for each test so they do
        not interact with each other.
        '''
        # pylint: disable=attribute-defined-outside-init
        self.image_name = f'{method.__name__}'

    def teardown_method(self, _method):
        '''Clean up containers and images that were left behind.'''
        clean_docker(self.image_name)

    def test_build_image(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest',
                              'RUN ["echo", "Hello, World!"]')

        uncap(dockerfile)

        image, logs = build_image(dockerfile, self.image_name)

        uncap(logs)

        assert image is not None
        assert image.tags == [f'{self.image_name}:latest']

        assert 'Successfully built' in logs
        assert f'Successfully tagged {self.image_name}:latest' in logs
        assert 'Hello, World!' in logs

    def test_build_image_args(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest',

                              # Test build args pass through
                              'ARG arg',
                              'RUN test "$arg" = "value"',

                              # Test files adjacent to the Dockerfile are part of the build context
                              'COPY some-file some-file',
                              'RUN test -f some-file')

        tmp_file('some-file', 'Hello!')

        build_args = {
            'arg': 'value'
        }

        image, log = build_image(dockerfile, self.image_name, build_args)

        uncap(log)

        assert image is not None
        assert 'Successfully built' in log
        assert f'Successfully tagged {self.image_name}:latest' in log

    def test_rebuild_image(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest',
                              'RUN ["echo", "Hello,"]')

        uncap(dockerfile)

        image, logs = build_image(dockerfile, self.image_name)

        uncap(logs)

        assert image is not None

        assert 'Successfully built' in logs
        assert 'Hello,' in logs
        assert 'World!' not in logs
        assert f'Successfully tagged {self.image_name}:latest' in logs

        with dockerfile.open('a') as f:
            f.write('RUN ["echo", "World!"]\n')

        image, logs = build_image(dockerfile, self.image_name)

        uncap(logs)

        assert 'Successfully built' in logs
        assert 'Hello,' in logs
        assert 'World!' in logs
        assert f'Successfully tagged {self.image_name}:latest' in logs

    def test_rebuild_image_change_buildargs(self, tmp_file, uncap):
        initial_uid = 30120
        initial_gid = 30121
        next_uid = 30122
        next_gid = 30123

        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM alpine:latest',
                              'ARG user_id',
                              'ARG group_id',
                              'RUN addgroup -g $group_id testuser \\',
                              '  && adduser -u $user_id -D -G testuser testuser',
                              'USER testuser',
                              'RUN echo "User: $(id -u)"',
                              'RUN echo "Group: $(id -g)"')

        uncap(dockerfile)

        build_args = {
            'user_id': f'{initial_uid}',
            'group_id': f'{initial_gid}',
        }

        image, logs = build_image(dockerfile, self.image_name, build_args)

        assert image is not None
        assert 'Successfully built' in logs
        assert f'Successfully tagged {self.image_name}:latest' in logs
        assert f'User: {initial_uid}' in logs
        assert f'Group: {initial_gid}' in logs
        assert f'User: {next_uid}' not in logs
        assert f'Group: {next_gid}' not in logs

        # Store the current ID because it is about to be replaced by a new image
        # with different build arguments. This is used to clean up the image
        # later.
        previous_image = image.id

        build_args = {
            'user_id': f'{next_uid}',
            'group_id': f'{next_gid}',
        }

        image, logs = build_image(dockerfile, self.image_name, build_args)

        assert image is not None
        assert 'Successfully built' in logs
        assert f'Successfully tagged {self.image_name}:latest' in logs
        assert f'User: {initial_uid}' not in logs
        assert f'Group: {initial_gid}' not in logs
        assert f'User: {next_uid}' in logs
        assert f'Group: {next_gid}' in logs

        assert previous_image is not None
        clean_docker(previous_image)

    def test_invalid_image(self, tmp_file, uncap):
        dockerfile = tmp_file('test-image.dockerfile',
                              'FROM invalid_image:latest')

        uncap(dockerfile)

        with pytest.raises(BuildError):
            build_image(dockerfile, self.image_name)

    def test_can_base_on_another_image(self, tmp_file, uncap):
        expected_uid = 30120
        expected_gid = 30121
        expected_uname = 'test-user'
        expected_gname = 'test-group'

        name_1 = f'{self.image_name}-1'
        name_2 = f'{self.image_name}-2'

        dockerfile_1 = tmp_file(f'{name_1}.dockerfile',
                                'FROM alpine:latest',
                                'ARG user_id',
                                'ARG user_name',
                                'ARG group_id',
                                'ARG group_name',
                                'RUN addgroup -g $group_id $group_name \\',
                                '  && adduser -u $user_id -D -G $group_name $user_name',
                                'USER $user_name')

        dockerfile_2 = tmp_file(f'{name_2}.dockerfile',
                                f'FROM {name_1}:latest',
                                f'RUN test "$(id -u)" = "{expected_uid}"',
                                f'RUN test "$(id -g)" = "{expected_gid}"',
                                f'RUN test "$(id -un)" = "{expected_uname}"',
                                f'RUN test "$(id -gn)" = "{expected_gname}"')

        uncap(dockerfile_1)
        uncap(dockerfile_2)

        build_args = {
            'user_id': expected_uid,
            'group_id': expected_gid,
            'user_name': expected_uname,
            'group_name': expected_gname,
        }

        try:
            image_1, logs_1 = build_image(dockerfile_1, name_1, build_args)
            image_2, logs_2 = build_image(dockerfile_2, name_2, build_args)
        finally:
            clean_docker(name_1)
            clean_docker(name_2)

        uncap(logs_1)
        uncap(logs_2)

        assert image_1 is not None
        assert image_2 is not None

        assert 'Successfully built' in logs_1
        assert 'Successfully tagged' in logs_1
        assert f'Successfully tagged {name_1}:latest' in logs_1
        assert name_2 not in logs_1

        assert 'Successfully built' in logs_2
        assert 'Successfully tagged' in logs_2
        assert f'Successfully tagged {name_2}:latest' in logs_2
        assert name_1 in logs_2
