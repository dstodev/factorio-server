'''Test the Start Command.'''

import json
import os
from functools import partial
from test.util_docker import clean_docker

from manage import game, paths
from manage.command import Backup, Download
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

    name = 'test-game-backup-no-files'

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
