'''Tests for build_args function'''

import json
from functools import partial

from manage import game_files, paths
from manage.docker.build_args import build_args


def test_build_args(tmp_path, mocker, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    game = 'test-game'

    cfg_file = game_files.cfg_file(game)
    cfg_file.parent.mkdir(parents=True, exist_ok=True)
    cfg_file.write_text(json.dumps({
        "port": {
            "game": 34197,
            "rcon": 34207
        },
        "user": {
            "name": "server-user:30121",
            "group": "server-group:30122"
        }
    }, indent=2))

    uncap(cfg_file)

    assert build_args(game) == {
        'game_port': '34197',
        'rcon_port': '34207',
        'user_name': 'server-user',
        'user_id': '30121',
        'group_name': 'server-group',
        'group_id': '30122'
    }


def test_build_args_only_ports(tmp_path, mocker, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    game = 'test-game'

    cfg_file = game_files.cfg_file(game)
    cfg_file.parent.mkdir(parents=True, exist_ok=True)
    cfg_file.write_text(json.dumps({
        "port": {
            "game": 34197,
            "rcon": 34207
        },
    }, indent=2))

    uncap(cfg_file)

    assert build_args(game) == {
        'game_port': '34197',
        'rcon_port': '34207',
    }


def test_build_args_only_user(tmp_path, mocker, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    game = 'test-game'

    cfg_file = game_files.cfg_file(game)
    cfg_file.parent.mkdir(parents=True, exist_ok=True)
    cfg_file.write_text(json.dumps({
        "user": {
            "name": "server-user:30121",
            "group": "server-group:30122"
        }
    }, indent=2))

    uncap(cfg_file)

    assert build_args(game) == {
        'user_name': 'server-user',
        'user_id': '30121',
        'group_name': 'server-group',
        'group_id': '30122'
    }
