'''Test RCON helpers.'''

import json
import shutil
from functools import partial

import pytest
from manage import game, paths, rcon
from manage.docker.container import GameContainer
from manage.docker.util import build_image as _build_image
from manage.rcon import RconError
from manage.util import clean_docker


# Patch build_image to disable caching for all tests in this module
def build_image(*args, **kwargs):
    kwargs.pop('cache', None)
    return _build_image(*args, **kwargs, cache=False)


def test_rcon_password(mocker, tmp_path, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    password = rcon.password(name, 10)

    password_file = game.cfg_dir(name) / 'secret'

    uncap(password_file)
    uncap(password)

    assert len(password) == 10
    assert all(c.isalnum() for c in password)

    assert password_file.is_file()
    assert password_file.read_text(encoding='utf-8') == f'{password}\n'
    assert password_file.stat().st_mode & 0o777 == 0o600


def test_rcon_password_persists(mocker, tmp_path):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    password = rcon.password(name, 10)

    assert password == rcon.password(name, 10)
    assert password == rcon.password(name, 10)


def test_rcon_password_new(mocker, tmp_path, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    password = rcon.password(name, 10)

    assert password == rcon.password(name, 10)

    new_password = rcon.password(name, 10, new=True)
    assert password != new_password

    uncap(game.cfg_dir(name) / 'secret')
    uncap(password)
    uncap(new_password)

    assert len(new_password) == 10
    assert all(c.isalnum() for c in new_password)

    assert new_password == rcon.password(name, 10)
    assert new_password == rcon.password(name, 10)


def test_rcon_password_length(mocker, tmp_path, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    password = rcon.password(name, 20)

    uncap(game.cfg_dir(name) / 'secret')
    uncap(password)

    assert len(password) == 20
    assert all(c.isalnum() for c in password)


def test_rcon_send_shim_echo(mocker, tmp_path, tmp_file, uncap):
    '''Use a shim script to echo passed args & stdin for validation.'''
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-rcon-stdin'

    rcon_shim = tmp_file(f'cfg/{name}/rcon.sh',
                         '#!/bin/sh',
                         'read -r password',
                         'echo "$password"',
                         'echo "$@"',
                         mode=0o755)

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          'FROM alpine:latest',
                          'COPY rcon.sh /usr/bin/rcon')

    server_json = tmp_file(f'cfg/{name}/config.json',
                           json.dumps({
                               'port': {
                                   'rcon': 12345
                               },
                           }, indent=2))

    uncap(rcon_shim)
    uncap(dockerfile)
    uncap(server_json)

    command = ['echo', 'Hello!']

    image, _logs = build_image(dockerfile, name)

    container = GameContainer(f'{name}-server', image)

    try:
        container.start(command=['tail', '-f', '/dev/null'])  # Run until manually stopped

        result = rcon.send(name, command)
        password = rcon.password(name)

        assert container.container is not None
        container.container.stop()

    finally:
        clean_docker(f'{name}-server')

    assert result is not None
    assert result.exit_status == 0
    assert result.output == f'{password}\n127.0.0.1:12345 echo Hello!\n'


def test_rcon_send_no_cfg(mocker, tmp_path):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    with pytest.raises(RconError, match=f'No RCON port found in configuration for game: {name}'):
        rcon.send(name, ['echo', 'Hello!'])


def test_rcon_send_no_container(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    tmp_file(f'cfg/{name}/config.json',
             json.dumps({
                 'port': {
                     'rcon': 12345
                 },
             }, indent=2))

    uncap(tmp_file)

    with pytest.raises(RconError, match='No target container'):
        rcon.send(name, ['echo', 'Hello!'])


def test_rcon_send_no_password(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    tmp_file(f'cfg/{name}/config.json',
             json.dumps({
                 'port': {
                     'rcon': 12345
                 },
             }, indent=2))

    uncap(tmp_file)


def test_rcon_send_no_client_in_image(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-rcon-send-no-client-in-image'

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          'FROM alpine:latest')

    server_json = tmp_file(f'cfg/{name}/config.json',
                           json.dumps({'port': {'rcon': 12345}}, indent=2))

    uncap(dockerfile)
    uncap(server_json)

    try:
        image, _logs = game.docker_image(name, cache=False)

        container = GameContainer(f'{name}-server', image)

        container.start(command=['tail', '-f', '/dev/null'])

        with pytest.raises(RconError) as e:
            rcon.send(name, ['echo', 'Hello!'])

        uncap(f'{e.value}'.strip())
        assert 'RCON client failure' f'{e.value}'
        assert 'not found' in f'{e.value}'
        assert 'rcon' in f'{e.value}'

        assert container.container is not None
        container.container.stop()
        container.wait()

        assert container.container is None

    finally:
        clean_docker(name)
        clean_docker(f'{name}-server')


def test_rcon_send(mocker, tmp_path, tmp_file, uncap):
    shutil.copytree(paths.get('rcon'), tmp_path / 'rcon')
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-rcon-send'
    rcon_image_name = 'rcon-test-rcon-send'

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          f'FROM {rcon_image_name}:latest',
                          'ARG user_id',
                          'ARG user_name',
                          'ARG group_id',
                          'ARG group_name',
                          'RUN groupadd - -gid "$group_id" "$group_name" \\',
                          '&& useradd - -uid "$user_id" \\',
                          '- -gid "$group_id" \\',
                          '- -create-home \\',
                          '"$user_name"')

    expected_uid = 30120
    expected_gid = 30121

    server_json = tmp_file(f'cfg/{name}/config.json',
                           json.dumps({
                               'port': {
                                   'rcon': 12345
                               },
                               'user': {
                                   'name': f'server-user:{expected_uid}',
                                   'group': f'server-group:{expected_gid}'
                               }
                           }, indent=2))

    uncap(dockerfile)
    uncap(server_json)

    try:
        build_image(paths.get('rcon') / 'Dockerfile', rcon_image_name, game.build_args(name))

        image, _logs = game.docker_image(name, cache=False)

        container = GameContainer(f'{name}-server', image)

        container.start(command=['tail', '-f', '/dev/null'])

        result = container.execute(['rcon', 'test'])

        assert result is not None
        assert result.exit_status == 0
        assert result.output == 'All tests passed!\n'

        result = container.execute(['/bin/sh', '-c', 'stat -c "%u:%g" "$(which rcon)"'])

        assert result is not None
        assert result.exit_status == 0
        assert result.output == '0:0\n'

        result = container.execute(['/bin/sh', '-c', 'echo "$(id -u):$(id -g)"'])

        assert result is not None
        assert result.exit_status == 0
        assert result.output == f'{expected_uid}:{expected_gid}\n'

        with pytest.raises(RconError) as e:
            rcon.send(name, ['something'])

        uncap(f'{e.value}'.strip())
        assert 'RCON client failure' in f'{e.value}'
        assert 'Connection refused' in f'{e.value}'

        mocker.patch('manage.rcon.password', return_value=None)

        with pytest.raises(RconError) as e:
            rcon.send(name, ['something'])

        uncap(f'{e.value}'.strip())
        assert 'RCON client failure' in f'{e.value}'
        assert 'Timed out waiting for password' in f'{e.value}'

        assert container.container is not None
        container.container.stop()
        container.wait()

        assert container.container is None

    finally:
        clean_docker(f'{name}-server')
        # clean_docker(rcon_image_name)
