'''Test the Start Command.'''

import json
import os
from functools import partial

from manage import PROJECT_NAME_SHORT, game, paths, rcon
from manage.command import Backup, Download, Rcon, Start
from manage.util import clean_docker, timestamp


def test_backup(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game-backup'

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

    server_json = tmp_file(f'cfg/{name}/config.json',
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

    rcon_shim = tmp_file(f'cfg/{name}/rcon.sh',
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

    rcon_port = 12345
    save_cmd_str = '/save'

    server_json = tmp_file(f'cfg/{name}/config.json',
                           json.dumps({
                               'port': {
                                   'rcon': rcon_port
                               },
                               'rcon': {
                                   'save-pre': f'{save_cmd_str}-pre',
                                   'save': save_cmd_str,
                                   'save-post': f'{save_cmd_str}-post',
                               }
                           }, indent=2))

    uncap(rcon_shim)
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

    assert backup.try_save_pre is not None
    assert backup.try_save_post is not None
    assert len(backup.try_save_pre.actions) == 2
    assert len(backup.try_save_post.actions) == 1

    rcon_password = rcon.password(name)

    save_cmd = backup.try_save_pre.actions[0]
    assert isinstance(save_cmd, Rcon)
    assert save_cmd.last_result is not None
    assert save_cmd.last_result.exit_status == 0
    assert save_cmd.last_result.output == f'{rcon_password}\n127.0.0.1:{rcon_port} {save_cmd_str}-pre\n'

    save_cmd = backup.try_save_pre.actions[1]
    assert isinstance(save_cmd, Rcon)
    assert save_cmd.last_result is not None
    assert save_cmd.last_result.exit_status == 0
    assert save_cmd.last_result.output == f'{rcon_password}\n127.0.0.1:{rcon_port} {save_cmd_str}\n'

    save_cmd = backup.try_save_post.actions[0]
    assert isinstance(save_cmd, Rcon)
    assert save_cmd.last_result is not None
    assert save_cmd.last_result.exit_status == 0
    assert save_cmd.last_result.output == f'{rcon_password}\n127.0.0.1:{rcon_port} {save_cmd_str}-post\n'

    logs = list(game.logs_dir(name).iterdir())
    assert len(logs) == 1
    log = logs[0]
    assert log.name.endswith('.log')

    uncap(log)

    content = log.read_text(encoding='utf-8')

    assert content == f'({PROJECT_NAME_SHORT}) closing logfile writer\n'
