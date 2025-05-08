'''Tests for build_args function'''

import json
import os
from functools import partial

from manage import game, paths
from manage.command import Download
from manage.util import clean_docker


def test_download(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game-download'

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          'FROM alpine:latest',
                          # Force unique cache for next layers;
                          # fixes frequent failures due to shared intermediate containers being deleted
                          f'RUN echo "{name}"',
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
                      'echo "$1" > "$server_dir/some-file"',
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

    server_hot = game.server_dir(name)

    uncap(server_hot)

    assert server_hot.is_dir()
    assert server_hot.stat().st_uid == expected_uid
    assert server_hot.stat().st_gid == expected_gid

    assert server_hot.parent.stat().st_uid == os.getuid()
    assert server_hot.parent.stat().st_gid == os.getgid()

    expected_server_file = server_hot / 'some-file'

    assert expected_server_file.is_file()
    assert expected_server_file.stat().st_uid == expected_uid
    assert expected_server_file.stat().st_gid == expected_gid
    assert expected_server_file.read_text() == '/game/hot\n'


def test_repeated_download_does_nothing(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-repeated-download-does-nothing'

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          'FROM alpine:latest',
                          # Force unique cache for next layers;
                          # fixes frequent failures due to shared intermediate containers being deleted
                          f'RUN echo "{name}"',
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
                      'echo "$1" >> "$server_dir/some-file"',
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

    try:
        download = Download(name)
        download.execute()
    finally:
        clean_docker(name)

    server_hot = game.server_dir(name)

    uncap(server_hot)

    assert server_hot.is_dir()
    assert server_hot.stat().st_uid == expected_uid
    assert server_hot.stat().st_gid == expected_gid

    assert server_hot.parent.stat().st_uid == os.getuid()
    assert server_hot.parent.stat().st_gid == os.getgid()

    expected_server_file = server_hot / 'some-file'

    assert expected_server_file.is_file()
    assert expected_server_file.stat().st_uid == expected_uid
    assert expected_server_file.stat().st_gid == expected_gid
    assert expected_server_file.read_text() == '/game/hot\n'
