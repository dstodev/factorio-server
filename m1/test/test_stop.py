'''Test the Stop Command.'''

import json
from functools import partial
from test.util import PKillTail

from manage import PROJECT_NAME_SHORT, game, paths, rcon
from manage.command import Rcon, Start, Stop
from manage.util import clean_docker


def test_stop(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game-stop'

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          'FROM alpine:latest')

    start = tmp_file(f'cfg/{name}/start.sh',
                     '#!/bin/sh',
                     'echo "Hello!"',
                     'sleep 0.1',
                     mode=0o744)

    uncap(dockerfile)
    uncap(start)

    try:
        start = Start(name)
        start.execute(auto_rm=False)

        stop = Stop(name)
        stop.execute()

    finally:
        clean_docker(f'{name}-server')

    result = stop.last_result

    assert result is not None
    assert result.exit_status == 0
    assert 'Hello!\n' in result.stdout
    assert result.stderr == ''

    server_hot = game.server_dir(name)

    uncap(server_hot)

    logs = list(game.logs_dir(name).iterdir())
    assert len(logs) == 1
    log = logs[0]
    assert log.name.endswith('.log')

    uncap(log)

    content = log.read_text(encoding='utf-8')

    assert content == f'{result.stdout}({PROJECT_NAME_SHORT}) closing logfile writer\n'


def test_stop_no_container(mocker, tmp_path):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game-stop-no-container'

    stop = Stop(name)
    stop.execute()

    result = stop.last_result

    assert result is None


def test_stop_tries_rcon_save_stop(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game-stop-rcon-save-stop'

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
                     'echo "Hello!"',
                     'tail -f /dev/null',
                     mode=0o744)

    rcon_port = 12345
    save_cmd_str = '/save'
    stop_cmd_str = '/stop'

    server_json = tmp_file(f'cfg/{name}/config.json',
                           json.dumps({
                               'port': {
                                   'rcon': rcon_port
                               },
                               'rcon': {
                                   'save-pre': f'{save_cmd_str}-pre',
                                   'save': save_cmd_str,
                                   'stop': stop_cmd_str
                               }
                           }, indent=2))

    uncap(rcon_shim)
    uncap(dockerfile)
    uncap(start)
    uncap(server_json)

    try:
        start = Start(name)
        start.execute(auto_rm=False)

        stop = Stop(name)

        # Hack: Add a pkill command as the final command after the save and stop commands
        #       to simulate the server stopping.
        stop.try_rcon.add_action(PKillTail(start.container))
        stop.execute()

    finally:
        clean_docker(f'{name}-server')

    result = stop.last_result

    assert result is not None
    assert result.exit_status == 143  # 128 + 15 (SIGTERM) from pkill
    assert 'Hello!\n' in result.stdout
    assert result.stderr == 'Terminated\n'

    assert stop.try_rcon is not None
    assert len(stop.try_rcon.actions) == 4

    rcon_password = rcon.password(name)

    save_cmd = stop.try_rcon.actions[0]
    assert isinstance(save_cmd, Rcon)
    assert save_cmd.last_result is not None
    assert save_cmd.last_result.exit_status == 0
    assert save_cmd.last_result.output == f'{rcon_password}\n127.0.0.1:{rcon_port} {save_cmd_str}-pre\n'

    save_cmd = stop.try_rcon.actions[1]
    assert isinstance(save_cmd, Rcon)
    assert save_cmd.last_result is not None
    assert save_cmd.last_result.exit_status == 0
    assert save_cmd.last_result.output == f'{rcon_password}\n127.0.0.1:{rcon_port} {save_cmd_str}\n'

    stop_cmd = stop.try_rcon.actions[2]
    assert isinstance(stop_cmd, Rcon)
    assert stop_cmd.last_result is not None
    assert stop_cmd.last_result.exit_status == 0
    assert stop_cmd.last_result.output == f'{rcon_password}\n127.0.0.1:{rcon_port} {stop_cmd_str}\n'

    server_hot = game.server_dir(name)

    uncap(server_hot)

    logs = list(game.logs_dir(name).iterdir())
    assert len(logs) == 1
    log = logs[0]
    assert log.name.endswith('.log')

    uncap(log)

    content = log.read_text(encoding='utf-8')

    assert content == f'{result.stdout}{result.stderr}({PROJECT_NAME_SHORT}) closing logfile writer\n'
