'''Test the Start Command.'''

import json
import os
from functools import partial
from test.util_docker import clean_docker

from manage import game, paths
from manage.command import Download, Start


def test_start(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game-start'

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          'FROM alpine:latest',
                          'ARG user_id',
                          'ARG user_name',
                          'ARG group_id',
                          'ARG group_name',
                          'RUN addgroup -g $group_id $group_name \\',
                          '  && adduser -u $user_id -D -G $group_name $user_name',
                          'USER $user_name')

    start = tmp_file(f'cfg/{name}/start.sh',
                     '#!/bin/sh',
                     'echo "Server started!"',
                     'echo "$1"',
                     'test -f "$1/some-file"',
                     mode=0o744)

    download = tmp_file(f'cfg/{name}/download.sh',
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
    uncap(start)
    uncap(download)
    uncap(server_json)

    try:
        download = Download(name)
        download.execute()

        start = Start(name)
        start.execute(auto_rm=False)

        result = start.container.wait()

    finally:
        clean_docker(name)

    assert result is not None
    assert result.exit_status == 0

    server_hot = game.server_dir(name)

    uncap(server_hot)

    log = game.logs_dir(name) / 'server.log'

    uncap(log)

    assert log.is_file()
    assert log.stat().st_uid == os.getuid()
    assert log.stat().st_gid == os.getgid()

    content = log.read_text(encoding='utf-8')
    assert 'Server started!\n' in content
    assert '/game/hot\n' in content
    assert 'closing logfile writer\n' in content
