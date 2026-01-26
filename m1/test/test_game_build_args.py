'''Tests for build_args function'''

import json
from functools import partial

from manage import game, paths


def test_build_args(tmp_path, mocker, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    cfg_file = game.cfg_file(name)
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

    assert game.build_args(name) == {
        'game_port': '34197',
        'rcon_port': '34207',
        'user_name': 'server-user',
        'user_id': '30121',
        'group_name': 'server-group',
        'group_id': '30122'
    }


def test_build_args_only_ports(tmp_path, mocker, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    cfg_file = game.cfg_file(name)
    cfg_file.parent.mkdir(parents=True, exist_ok=True)
    cfg_file.write_text(json.dumps({
        "port": {
            "game": 34197,
            "rcon": 34207
        },
    }, indent=2))

    uncap(cfg_file)

    assert game.build_args(name) == {
        'game_port': '34197',
        'rcon_port': '34207',
    }


def test_build_args_only_user(tmp_path, mocker, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    cfg_file = game.cfg_file(name)
    cfg_file.parent.mkdir(parents=True, exist_ok=True)
    cfg_file.write_text(json.dumps({
        "user": {
            "name": "server-user:30121",
            "group": "server-group:30122"
        }
    }, indent=2) + '\n')

    uncap(cfg_file)

    assert game.build_args(name) == {
        'user_name': 'server-user',
        'user_id': '30121',
        'group_name': 'server-group',
        'group_id': '30122'
    }


def test_build_args_no_file(tmp_path, mocker):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))
    name = 'test-game'
    assert not game.build_args(name)
