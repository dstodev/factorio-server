'''Test the Shelf Command.'''

import json
import os
from functools import partial
from test.util_docker import clean_docker

from manage import paths
from manage.command import Download, Shelf


def test_shelf(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'
    server_dir = tmp_path / f'server-files/{name}'

    some_file = tmp_file(f'server-files/{name}/hot/some-file.txt',
                         'Hello!')

    uncap(some_file)

    assert server_dir.exists()

    shelf = Shelf(name)
    shelf.execute()

    expected_shelf_dir = tmp_path / f'shelf/{name}/1'
    expected_file = expected_shelf_dir / 'hot/some-file.txt'

    assert expected_shelf_dir.is_dir()
    assert expected_file.is_file()
    assert expected_file.read_text() == 'Hello!\n'
    assert not some_file.is_file()  # File was moved, not copied
    assert not server_dir.exists()  # Server dir was moved
    assert server_dir.parent.exists()


def test_shelf_twice(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'
    server_dir = tmp_path / f'server-files/{name}'

    count = 0

    def do():
        nonlocal count
        count += 1

        some_file = tmp_file(f'server-files/{name}/hot/some-file.txt',
                             f'Count: {count}')

        uncap(some_file)

        assert server_dir.exists()

        shelf = Shelf(name)
        shelf.execute()

        expected_shelf_dir = tmp_path / f'shelf/{name}/{count}'
        expected_file = expected_shelf_dir / 'hot/some-file.txt'

        uncap(expected_file)

        assert expected_shelf_dir.is_dir()
        assert expected_file.is_file()
        assert expected_file.read_text() == f'Count: {count}\n'
        assert not some_file.is_file()
        assert not server_dir.exists()
        assert server_dir.parent.exists()

    do()
    do()


def test_shelf_permissions(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game-shelf-permissions'

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          'FROM alpine:latest',
                          'ARG user_id',
                          'ARG user_name',
                          'ARG group_id',
                          'ARG group_name',
                          'RUN addgroup -g $group_id $group_name \\',
                          '  && adduser -u $user_id -D -G $group_name $user_name',
                          'USER $user_name')

    script = tmp_file(f'cfg/{name}/download.sh',
                      '#!/bin/sh',
                      'server_dir="$1"',
                      'touch "$server_dir/some-file"',
                      mode=0o744)

    expected_uid = 30120
    expected_gid = 30121

    server_json = tmp_file(f'cfg/{name}/server.json',
                           json.dumps({
                               'user': {
                                   'name': f'server-user:{expected_uid}',
                                   'group': f'server-group:{expected_gid}'
                               }
                           }, indent=2))

    uncap(dockerfile)
    uncap(script)
    uncap(server_json)

    try:
        download = Download(name)
        download.execute()
    finally:
        clean_docker(name)

    shelf = Shelf(name)
    shelf.execute()

    expected_shelf_dir = tmp_path / f'shelf/{name}/1'

    assert expected_shelf_dir.is_dir()
    assert expected_shelf_dir.stat().st_uid == os.getuid()
    assert expected_shelf_dir.stat().st_gid == os.getgid()

    assert (expected_shelf_dir / 'hot').is_dir()
    assert (expected_shelf_dir / 'hot').stat().st_uid == expected_uid
    assert (expected_shelf_dir / 'hot').stat().st_gid == expected_gid

    assert (expected_shelf_dir / 'hot/some-file').is_file()
    assert (expected_shelf_dir / 'hot/some-file').stat().st_uid == expected_uid
    assert (expected_shelf_dir / 'hot/some-file').stat().st_gid == expected_gid
