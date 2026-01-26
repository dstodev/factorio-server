'''Test the Start Command.'''

import json
import os
from functools import partial

from manage import PROJECT_NAME_SHORT, game, paths, rcon
from manage.command import Download, Start
from manage.util import clean_docker


def test_start(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game-start'

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

    start = tmp_file(f'cfg/{name}/start.sh',
                     '#!/bin/sh',
                     'echo "Server started!"',
                     'echo "$1"',
                     'echo "$2"',
                     'test -f "$1/some-file"',
                     'touch /tmp/stopfile',
                     mode=0o744)

    download = tmp_file(f'cfg/{name}/download.sh',
                        '#!/bin/sh',
                        'server_dir="$1"',
                        'touch "$server_dir/some-file"',
                        mode=0o744)

    expected_uid = 30120
    expected_gid = 30121

    server_json = tmp_file(f'cfg/{name}/config.json',
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
        clean_docker(f'{name}-download')
        clean_docker(f'{name}-server')

    assert result is not None
    assert result.exit_status == 0

    server_hot = game.server_dir(name)

    uncap(server_hot)

    logs = list(game.logs_dir(name).iterdir())
    assert len(logs) == 2
    log = logs[0]
    assert log.name.endswith('.log')

    uncap(log)

    assert log.is_file()
    assert log.stat().st_uid == os.getuid()
    assert log.stat().st_gid == os.getgid()

    content = log.read_text(encoding='utf-8')

    assert 'Server started!\n' in content
    assert '/game/hot\n' in content
    assert rcon.password(name) in content
    assert 'closing logfile writer\n' in content


def test_start_stopfile_loop(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game-start-stopfile-loop'

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          'FROM alpine:latest')

    start = tmp_file(f'cfg/{name}/start.sh',
                     '#!/bin/sh',
                     'if [ -f /tmp/call-count ]; then',
                     '  count=$(cat /tmp/call-count)',
                     '  count=$((count + 1))',
                     'else',
                     '  count=1',
                     'fi',
                     'echo "$count"',
                     'echo $count > /tmp/call-count',
                     'if [ $count -eq 10 ]; then',
                     '  touch /tmp/stopfile',
                     'fi',
                     mode=0o744)

    uncap(dockerfile)
    uncap(start)

    try:
        start = Start(name)
        start.execute(auto_rm=False)

        result = start.container.wait()

    finally:
        clean_docker(f'{name}-server')

    assert result is not None
    assert result.exit_status == 0

    server_hot = game.server_dir(name)

    uncap(server_hot)

    logs = list(game.logs_dir(name).iterdir())
    assert len(logs) == 1
    log = logs[0]
    assert log.name.endswith('.log')

    uncap(log)

    content = log.read_text(encoding='utf-8')

    assert content == f'1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n({PROJECT_NAME_SHORT}) closing logfile writer\n'
