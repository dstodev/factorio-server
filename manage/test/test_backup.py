'''Test the Start Command.'''

import json
import os
from functools import partial
from test.util import clean_docker

from manage import PROJECT_NAME, game, paths
from manage.command import Backup, Download, Rcon, Start
from manage.util import timestamp


def test_backup(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game-backup'

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          'FROM alpine:latest',
                          'ARG user_id',
                          'ARG user_name',
                          'ARG group_id',
                          'ARG group_name',
                          'RUN addgroup -g $group_id $group_name \\',
                          '  && adduser -u $user_id -D -G $group_name $user_name',
                          'USER $user_name')

    backup = tmp_file(f'cfg/{name}/backup.sh',
                      '#!/bin/sh',
                      'echo "$1"',
                      'echo "$2"',
                      'cp -r "$1" "$2"',
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
    uncap(backup)
    uncap(download)
    uncap(server_json)

    try:
        download = Download(name)
        download.execute()

        before_backup = timestamp()

        backup = Backup(name)
        backup.execute()

    finally:
        clean_docker(name)

    after_backup = timestamp()

    backup_dir = game.backup_dir(name)

    backups = list(backup_dir.iterdir())
    assert len(backups) == 1
    backup = backups[0]

    uncap(backup)

    assert backup.is_dir()
    assert backup.stat().st_uid == os.getuid()
    assert backup.stat().st_gid == os.getgid()

    assert (backup / 'hot').is_dir()
    assert (backup / 'hot').stat().st_uid == expected_uid
    assert (backup / 'hot').stat().st_gid == expected_gid

    assert (backup / 'hot/some-file').is_file()
    assert (backup / 'hot/some-file').stat().st_uid == expected_uid
    assert (backup / 'hot/some-file').stat().st_gid == expected_gid

    date = backup.name
    assert date >= before_backup
    assert date <= after_backup


def test_backup_no_files(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-backup-no-files'

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          'FROM alpine:latest')

    backup = tmp_file(f'cfg/{name}/backup.sh',
                      '#!/bin/sh',
                      'cp -r "$1" "$2"',
                      mode=0o744)

    uncap(dockerfile)
    uncap(backup)

    try:
        backup = Backup(name)
        backup.execute()

    finally:
        clean_docker(f'{name}-backup')

    backup_dir = game.backup_dir(name)

    assert not backup_dir.exists()

    game.server_dir(name).mkdir(parents=True, exist_ok=True)

    try:
        backup = Backup(name)
        backup.execute()
    finally:
        clean_docker(f'{name}-backup')

    assert not backup_dir.exists()


def test_backup_tries_rcon_save(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-backup-tries-rcon-save'

    rcon = tmp_file(f'cfg/{name}/rcon.sh',
                    '#!/bin/sh',
                    'read -r password',
                    'echo "$password"',
                    'echo "$@"',
                    mode=0o755)

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          'FROM alpine:latest',
                          'COPY rcon.sh /usr/bin/rcon')

    start = tmp_file(f'cfg/{name}/start.sh',
                     '#!/bin/sh',
                     'tail -f /dev/null',
                     mode=0o744)

    backup = tmp_file(f'cfg/{name}/backup.sh',
                      '#!/bin/sh',
                      'echo "Hello!"',
                      mode=0o744)

    tmp_file(f'server-files/{name}/hot/some-file')

    save_cmd_str = '/save'

    server_json = tmp_file(f'cfg/{name}/server.json',
                           json.dumps({
                               'rcon': {
                                   'save': save_cmd_str
                               }
                           }, indent=2))

    uncap(rcon)
    uncap(dockerfile)
    uncap(start)
    uncap(backup)
    uncap(server_json)

    try:
        backup = Backup(name)
        backup.execute()

    finally:
        clean_docker(f'{name}-backup')

    result = backup.last_result

    assert result is not None
    assert result.exit_status == 0
    assert result.stdout == 'Hello!\n'
    assert result.stderr == ''

    assert backup.last_save is None  # server was not running

    try:
        start = Start(name)
        start.execute(auto_rm=False)

        backup = Backup(name)
        backup.execute()

        assert start.container.container is not None
        start.container.container.stop()
        start.container.wait()

    finally:
        clean_docker(f'{name}-server')
        clean_docker(f'{name}-backup')

    result = backup.last_result

    assert result is not None
    assert result.exit_status == 0
    assert result.stdout == 'Hello!\n'
    assert result.stderr == ''

    rcon_password = game.rcon_password(name)

    save_cmd = backup.last_save

    assert isinstance(save_cmd, Rcon)
    assert save_cmd.last_result is not None
    assert save_cmd.last_result.exit_status == 0
    assert save_cmd.last_result.output == f'{rcon_password}\n{save_cmd_str}\n'

    server_hot = game.server_dir(name)

    uncap(server_hot)

    log = game.logs_dir(name) / 'server.log'

    uncap(log)

    content = log.read_text(encoding='utf-8')

    assert content == f'({PROJECT_NAME}) closing logfile writer\n'
