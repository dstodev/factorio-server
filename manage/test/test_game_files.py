'''Tests for GameFiles class and related functions in manage/game_files.py'''

import json
from functools import partial

import pytest

from manage import game_files, paths


def test_force_dir_creates_dir(tmp_path):
    target_dir = tmp_path / 'test_dir'
    assert not target_dir.exists(), f'Directory {target_dir} should not exist before call.'

    path = game_files.force_dir(target_dir)

    assert path == target_dir, f'Expected {target_dir}, received {path}'
    assert target_dir.is_dir(), f'Directory {target_dir} was not created.'


def test_force_dir_does_not_recreate_existing_dir(tmp_path):
    target_dir = tmp_path / 'test_dir'
    target_dir.mkdir(parents=True, exist_ok=True)
    target_file = target_dir / 'test_file.txt'
    target_file.touch()
    assert target_file.is_file(), f'File {target_file} should exist before call.'

    path = game_files.force_dir(target_dir)

    assert path == target_dir, f'Expected {target_dir}, received {path}'
    assert target_file.is_file(), f'File {target_file} should still exist after call.'


@pytest.mark.parametrize('func_name,expected_dir', [
    ('cfg_dir', 'cfg'),
    ('backup_dir', 'backup'),
    ('server_dir', 'server-files'),
    ('shelf_dir', 'shelf'),
])
def test_game_dirs(tmp_path, mocker, func_name, expected_dir):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    game = 'test-game'

    expected_cfg = tmp_path / expected_dir / game

    result = getattr(game_files, func_name)(game)

    assert result == expected_cfg, f'Expected {expected_cfg}, received {result}'


@pytest.mark.parametrize('func_name,expected_stem', [
    ('cfg_file', 'server.json'),
    ('dockerfile', 'server.dockerfile'),
    ('backup_script', 'backup.sh'),
    ('download_script', 'download.sh'),
    ('restore_script', 'restore.sh'),
    ('start_script', 'start.sh'),
])
def test_game_files(tmp_path, mocker, func_name, expected_stem):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    game = 'test-game'

    cfg_dir = tmp_path / 'cfg' / game
    expected_cfg_file = cfg_dir / expected_stem

    result = getattr(game_files, func_name)(game)

    assert result == expected_cfg_file, f'Expected {expected_cfg_file}, received {result}'


def test_cfg_data(tmp_path, mocker):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    game = 'test-game'

    expected_data = {'key': 'value'}

    mocker.patch('builtins.open', mocker.mock_open(read_data=json.dumps(expected_data)))

    result = game_files.cfg_data(game)

    assert result == expected_data, f'Expected {expected_data}, received {result}'
