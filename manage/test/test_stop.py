'''Test the Stop Command.'''

import json
from functools import partial
from test.util_docker import clean_docker

from manage import PROJECT_NAME, game, paths
from manage.command import Rcon, Start, Stop
from manage.docker import GameContainer


class PkillTail:
    '''Hack: Add a pkill command as the final command after the save and stop commands
    to simulate the server stopping.'''

    def __init__(self, container: GameContainer):
        self.container = container

    def execute(self):
        self.container.execute(['pkill', 'tail'])


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

    log = game.logs_dir(name) / 'server.log'

    uncap(log)

    content = log.read_text(encoding='utf-8')

    assert content == f'{result.stdout}({PROJECT_NAME}) closing logfile writer\n'


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
                     'echo "Hello!"',
                     'tail -f /dev/null',
                     mode=0o744)

    save_cmd_str = '/save'
    stop_cmd_str = '/stop'

    server_json = tmp_file(f'cfg/{name}/server.json',
                           json.dumps({
                               'rcon': {
                                   'save': save_cmd_str,
                                   'stop': stop_cmd_str
                               }
                           }, indent=2))

    uncap(rcon)
    uncap(dockerfile)
    uncap(start)
    uncap(server_json)

    try:
        start = Start(name)
        start.execute(auto_rm=False)

        stop = Stop(name)
        stop.execute(add_cmds=[PkillTail(start.container)])

    finally:
        clean_docker(f'{name}-server')

    result = stop.last_result

    assert result is not None
    assert result.exit_status == 143  # 128 + 15 (SIGTERM) from pkill
    assert 'Hello!\n' in result.stdout
    assert result.stderr == 'Terminated\n'

    assert stop.last_schedule is not None
    assert len(stop.last_schedule.actions) == 3

    rcon_password = game.rcon_password(name)

    save_cmd = stop.last_schedule.actions[0]
    assert isinstance(save_cmd, Rcon)
    assert save_cmd.last_result is not None
    assert save_cmd.last_result.exit_status == 0
    assert save_cmd.last_result.output == f'{rcon_password}\n{save_cmd_str}\n'

    stop_cmd = stop.last_schedule.actions[1]
    assert isinstance(stop_cmd, Rcon)
    assert stop_cmd.last_result is not None
    assert stop_cmd.last_result.exit_status == 0
    assert stop_cmd.last_result.output == f'{rcon_password}\n{stop_cmd_str}\n'

    server_hot = game.server_dir(name)

    uncap(server_hot)

    log = game.logs_dir(name) / 'server.log'

    uncap(log)

    content = log.read_text(encoding='utf-8')

    assert content == f'{result.stdout}{result.stderr}({PROJECT_NAME}) closing logfile writer\n'
