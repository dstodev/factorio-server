import json
from functools import partial

from manage import game, paths, rcon
from manage.docker.container import GameContainer
from manage.docker.util import build_image
from manage.util import clean_docker


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


def test_rcon_stdin(mocker, tmp_path, tmp_file, uncap):
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

    server_json = tmp_file(f'cfg/{name}/server.json',
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
