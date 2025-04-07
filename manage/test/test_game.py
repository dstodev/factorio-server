import json
from functools import partial

import pytest

from manage import paths
from manage.game import Game, assert_file, force_dir


def test_force_dir_creates_dir(tmp_path):
    target_dir = tmp_path / 'test_dir'
    assert not target_dir.exists(), f'Directory {target_dir} should not exist before call.'

    dir = force_dir(target_dir)

    assert dir == target_dir, f'Expected {target_dir}, received {dir}'
    assert target_dir.is_dir(), f'Directory {target_dir} was not created.'


def test_force_dir_does_not_recreate_existing_dir(tmp_path):
    target_dir = tmp_path / 'test_dir'
    target_dir.mkdir(parents=True, exist_ok=True)
    target_file = target_dir / 'test_file.txt'
    target_file.touch()
    assert target_file.is_file(), f'File {target_file} should exist before call.'

    dir = force_dir(target_dir)

    assert dir == target_dir, f'Expected {target_dir}, received {dir}'
    assert target_file.is_file(), f'File {target_file} should still exist after call.'


def test_assert_file_failure(tmp_path):
    target_file = tmp_path / 'non_existent_file.txt'
    assert not target_file.exists(), f'File {target_file} should not exist before call.'

    with pytest.raises(AssertionError):
        assert_file(target_file)


def test_assert_file_success(tmp_path):
    target_file = tmp_path / 'test_file.txt'
    target_file.touch()
    assert target_file.is_file(), f'File {target_file} should exist before call.'

    dir = assert_file(target_file)

    assert dir == target_file, f'Expected {target_file}, received {dir}'


@pytest.mark.parametrize('method,expected_dir', [
    ('cfg_dir', 'cfg'),
    ('backup_dir', 'backup'),
    ('shelf_dir', 'shelf'),
])
def test_game_dirs(tmp_path, mocker, method, expected_dir):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test_game'
    game = Game(name)

    expected_cfg = paths.get(expected_dir) / name

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
    game = Game(name)

    cfg_dir = game.cfg_dir()

    expected_cfg_file = cfg_dir / expected_stem
    expected_cfg_file.touch()  # pass assertion for file existence

    result = getattr(game, method_name)()

    assert result == expected_cfg_file, f'Expected {expected_cfg_file}, received {result}'


def test_cfg_data(tmp_path, mocker):
    mocker.patch('manage.game.assert_file', side_effect=lambda f: f)
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test_game'
    game = Game(name)

    expected_data = {'key': 'value'}

    mocker.patch('builtins.open', mocker.mock_open(read_data=json.dumps(expected_data)))

    result = game.cfg_data()

    assert result == expected_data, f'Expected {expected_data}, received {result}'
