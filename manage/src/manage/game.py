'''Game-specific file and directory tools.'''

import json
import shutil
import tempfile
from pathlib import Path
from typing import NamedTuple

from docker.models.images import Image

from manage import paths
from manage.docker.util import build_image


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
    path = paths.get('server-hot') / name / 'hot'
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


def docker_image(name: str, context_files: list[Path] | None = None) -> tuple[Image, str]:
    '''Build the game's image.

    Build errors are raised as exceptions.

    :param name: The name of the game.
    :type name: str

    :param context_files: Additional files to include in the Docker context.
    :type context_files: list[Path] | None

    :return: The image and logs from the build.
    :rtype: tuple[Image, str]
    '''
    with tempfile.TemporaryDirectory() as tmpdir:
        dockerfile_ = docker_context(dockerfile(name), Path(tmpdir), context_files)
        image, logs = build_image(dockerfile_, name, build_args(name))

    return image, logs


def docker_context(dockerfile_: Path,
                   context_dir: Path,
                   context_files: list[Path] | None = None) -> Path:
    '''Prepare the Docker context inside of context_dir.

    Creates copies in context_dir of all files around the dockerfile & those
    listed in context_files. Unfortunately, lightweight symlinks cannot be used
    instead, because Docker will not use them.

    Returns a Path to the dockerfile in the context_dir.
    '''
    assert context_dir.is_dir(), f'Context directory must exist: {context_dir}'

    context_files = context_files or []

    for file in dockerfile_.parent.iterdir():
        copy_file(file, context_dir / file.name)

    for file in context_files:
        copy_file(file, context_dir / file.name)

    return context_dir / dockerfile_.name


def copy_file(src: Path, dst: Path):
    '''Copy a file or directory.'''
    if src.is_dir():
        shutil.copytree(src, dst)
    else:
        shutil.copy(src, dst)


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
