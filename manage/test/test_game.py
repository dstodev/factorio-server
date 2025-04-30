'''Tests for GameFiles class and related functions in manage/game.py'''

import json
from functools import partial

import pytest

from manage import game, paths


def test_force_dir_creates_dir(tmp_path):
    target_dir = tmp_path / 'test_dir'
    assert not target_dir.exists()

    path = game.force_dir(target_dir)

    assert path == target_dir
    assert target_dir.is_dir()


def test_force_dir_does_not_recreate_existing_dir(tmp_path):
    target_dir = tmp_path / 'test_dir'
    target_dir.mkdir(parents=True, exist_ok=True)
    target_file = target_dir / 'test_file.txt'
    target_file.touch()
    assert target_file.is_file()

    path = game.force_dir(target_dir)

    assert path == target_dir
    assert target_file.is_file()


@pytest.mark.parametrize('func_name,expected_dir', [
    ('cfg_dir', 'cfg'),
    ('backup_dir', 'backup'),
    ('server_dir', 'server-files'),
    ('shelf_dir', 'shelf'),
])
def test_game_dirs(tmp_path, mocker, func_name, expected_dir):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    expected_cfg = tmp_path / expected_dir / name

    if func_name == 'server_dir':
        expected_cfg = expected_cfg / 'hot'

    result = getattr(game, func_name)(name)

    assert result == expected_cfg


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

    name = 'test-game'

    cfg_dir = tmp_path / 'cfg' / name
    expected_cfg_file = cfg_dir / expected_stem

    result = getattr(game, func_name)(name)

    assert result == expected_cfg_file


def test_cfg_data(tmp_path, mocker):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    expected_data = {'key': 'value'}

    mocker.patch('builtins.open', mocker.mock_open(read_data=json.dumps(expected_data)))

    result = game.cfg_data(name)

    assert result == expected_data


def test_docker_context(tmp_path):
    docker_dir = tmp_path / 'docker'
    docker_dir.mkdir()

    dockerfile = docker_dir / 'server.dockerfile'
    dockerfile.touch()

    (docker_dir / 'file.txt').touch()  # related by adjacency to dockerfile

    context_dir = tmp_path / 'context'
    context_dir.mkdir()

    dockerfile = game.docker_context(dockerfile, context_dir)

    assert len(list(dockerfile.parent.iterdir())) == 2

    assert (context_dir / 'server.dockerfile').exists()
    assert (context_dir / 'file.txt').exists()


def test_docker_context_extra_files(tmp_path):
    docker_dir = tmp_path / 'docker'
    docker_dir.mkdir()

    dockerfile = docker_dir / 'server.dockerfile'
    dockerfile.touch()

    unrelated_dir = tmp_path / 'unrelated'
    unrelated_dir.mkdir()

    unrelated_file = unrelated_dir / 'file.txt'
    unrelated_file.touch()

    context_dir = tmp_path / 'context'
    context_dir.mkdir()

    # Add unrelated_file twice; once by adding its directory, and again by adding it directly.
    dockerfile = game.docker_context(dockerfile, context_dir, context_files=[unrelated_dir, unrelated_file])

    assert dockerfile.parent == context_dir
    assert len(list(dockerfile.parent.iterdir())) == 3
    assert (context_dir / 'server.dockerfile').exists()
    assert (context_dir / 'unrelated/file.txt').exists()  # copy directory
    assert (context_dir / 'file.txt').exists()  # copy file
