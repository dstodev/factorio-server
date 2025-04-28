'''Game-specific file and directory tools.'''

import json
from pathlib import Path
from typing import NamedTuple

from manage import paths


def force_dir(path: Path, **kwargs) -> Path:
    '''Get the path to a directory, creating it if it doesn't exist.'''
    path.mkdir(parents=True, exist_ok=True, **kwargs)
    assert path.is_dir(), f'Directory {path} does not exist.'
    return path


def cfg_dir(name: str) -> Path:
    '''Get the path to the game's configuration directory.'''
    path = paths.get('cfg') / name
    return path


def backup_dir(name: str) -> Path:
    '''Get the path to the game's backup directory.'''
    path = paths.get('backup') / name
    return path


def server_dir(name: str) -> Path:
    '''Get the path to the game's runtime directory.
    The server may be actively running here.
    '''
    path = paths.get('server-hot') / name
    return path


def shelf_dir(name: str) -> Path:
    '''Get the path to the game's shelf directory.'''
    path = paths.get('shelf') / name
    return path


def cfg_file(name: str) -> Path:
    '''Get the path to the game's configuration file.'''
    path = cfg_dir(name) / 'server.json'
    return path


def cfg_data(name: str) -> dict:
    '''Get the game's configuration data.'''

    cfg = cfg_file(name)
    data = {}

    with open(cfg, 'r', encoding='utf-8') as file:
        data = json.load(file)

    return data


def dockerfile(name: str) -> Path:
    '''Get the path to the game server Dockerfile.'''
    dockerfile_ = cfg_dir(name) / 'server.dockerfile'
    return dockerfile_


def backup_script(name: str) -> Path:
    '''Get the path to the server backup script.'''
    script = cfg_dir(name) / 'backup.sh'
    return script


def download_script(name: str) -> Path:
    '''Get the path to the server download script.'''
    script = cfg_dir(name) / 'download.sh'
    return script


def restore_script(name: str) -> Path:
    '''Get the path to the server restore script.'''
    script = cfg_dir(name) / 'restore.sh'
    return script


def start_script(name: str) -> Path:
    '''Get the path to the server start script.'''
    script = cfg_dir(name) / 'start.sh'
    return script


def build_args(name: str) -> dict[str, str]:
    '''Return common build arguments for game dockerfiles.

    Dockerfiles commonly include the lines:

    .. code-block:: dockerfile
        ARG game_port
        ARG rcon_port
        ARG user_id
        ARG user_name
        ARG group_id
        ARG group_name
    '''

    args = {}

    cfg = cfg_data(name)

    try:
        fields = {}
        ports = cfg['port']
        fields['game_port'] = str(ports['game'])
        fields['rcon_port'] = str(ports['rcon'])
        args.update(fields)
    except KeyError:
        pass

    try:
        fields = {}
        user_ = user(name)
        fields['user_id'] = str(user_.uid)
        fields['user_name'] = user_.name
        fields['group_id'] = str(user_.gid)
        fields['group_name'] = user_.group
        args.update(fields)
    except KeyError:
        pass

    return args


class User(NamedTuple):
    '''User information.'''
    name: str
    uid: int
    group: str
    gid: int


def user(name: str) -> User:
    '''Get the user and group information for the game server.'''
    user_ = cfg_data(name)['user']
    name, uid = user_['name'].split(':')
    group, gid = user_['group'].split(':')
    return User(name=name, uid=int(uid), group=group, gid=int(gid))
