'''Game-specific file and directory tools.'''

import json
from pathlib import Path

from manage import paths


def force_dir(path: Path, **kwargs) -> Path:
    '''Get the path to a directory, creating it if it doesn't exist.'''
    path.mkdir(parents=True, exist_ok=True, **kwargs)
    assert path.is_dir(), f'Directory {path} does not exist.'
    return path


def cfg_dir(game: str) -> Path:
    '''Get the path to the game's configuration directory.'''
    path = paths.get('cfg') / game
    return path


def backup_dir(game: str) -> Path:
    '''Get the path to the game's backup directory.'''
    path = paths.get('backup') / game
    return path


def server_dir(game: str) -> Path:
    '''Get the path to the game's runtime directory.
    The server may be actively running here.
    '''
    path = paths.get('server-hot') / game
    return path


def shelf_dir(game: str) -> Path:
    '''Get the path to the game's shelf directory.'''
    path = paths.get('shelf') / game
    return path


def cfg_file(game: str) -> Path:
    '''Get the path to the game's configuration file.'''
    path = cfg_dir(game) / 'server.json'
    return path


def cfg_data(game: str) -> dict:
    '''Get the game's configuration data.'''

    cfg = cfg_file(game)
    data = {}

    with open(cfg, 'r', encoding='utf-8') as file:
        data = json.load(file)

    return data


def dockerfile(game: str) -> Path:
    '''Get the path to the game server Dockerfile.'''
    dockerfile_ = cfg_dir(game) / 'server.dockerfile'
    return dockerfile_


def backup_script(game: str) -> Path:
    '''Get the path to the server backup script.'''
    script = cfg_dir(game) / 'backup.sh'
    return script


def download_script(game: str) -> Path:
    '''Get the path to the server download script.'''
    script = cfg_dir(game) / 'download.sh'
    return script


def restore_script(game: str) -> Path:
    '''Get the path to the server restore script.'''
    script = cfg_dir(game) / 'restore.sh'
    return script


def start_script(game: str) -> Path:
    '''Get the path to the server start script.'''
    script = cfg_dir(game) / 'start.sh'
    return script
