'''Tests for GameFiles class and related functions in manage/game_files.py'''

import json
from functools import partial

import pytest

from manage import paths
from manage.game_files import GameFiles, force_dir


def test_force_dir_creates_dir(tmp_path):
    target_dir = tmp_path / 'test_dir'
    assert not target_dir.exists(), f'Directory {target_dir} should not exist before call.'

    path = force_dir(target_dir)

    assert path == target_dir, f'Expected {target_dir}, received {path}'
    assert target_dir.is_dir(), f'Directory {target_dir} was not created.'


def test_force_dir_does_not_recreate_existing_dir(tmp_path):
    target_dir = tmp_path / 'test_dir'
    target_dir.mkdir(parents=True, exist_ok=True)
    target_file = target_dir / 'test_file.txt'
    target_file.touch()
    assert target_file.is_file(), f'File {target_file} should exist before call.'

    path = force_dir(target_dir)

    assert path == target_dir, f'Expected {target_dir}, received {path}'
    assert target_file.is_file(), f'File {target_file} should still exist after call.'


@pytest.mark.parametrize('method,expected_dir', [
    ('server_dir', 'server-files'),
    ('cfg_dir', 'cfg'),
    ('backup_dir', 'backup'),
    ('shelf_dir', 'shelf'),
])
def test_game_dirs(tmp_path, mocker, method, expected_dir):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test_game'
    game = GameFiles(name)

    expected_cfg = tmp_path / expected_dir / name

    result = getattr(game, method)()

    assert result == expected_cfg, f'Expected {expected_cfg}, received {result}'
    assert result.is_dir(), f'Directory {result} must exist.'


@pytest.mark.parametrize('method_name,expected_stem', [
    ('cfg_file', 'test_game.json'),
    ('dockerfile', 'test_game.dockerfile'),
    ('start_script', 'start.sh'),
    ('download_script', 'download.sh'),
    ('backup_script', 'backup.sh'),
    ('restore_script', 'restore.sh'),
])
def test_game_files(tmp_path, mocker, method_name, expected_stem):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test_game'
    game = GameFiles(name)

    cfg_dir = game.cfg_dir()

    expected_cfg_file = cfg_dir / expected_stem
    expected_cfg_file.touch()  # pass assertion for file existence

    result = getattr(game, method_name)()

    assert result == expected_cfg_file, f'Expected {expected_cfg_file}, received {result}'


def test_cfg_data(tmp_path, mocker):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test_game'
    game = GameFiles(name)

    expected_data = {'key': 'value'}

    mocker.patch('builtins.open', mocker.mock_open(read_data=json.dumps(expected_data)))

    result = game.cfg_data()

    assert result == expected_data, f'Expected {expected_data}, received {result}'
