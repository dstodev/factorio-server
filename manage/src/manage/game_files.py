'''Game-specific file and directory tools.'''

import json
import pathlib

from manage import paths


def force_dir(path: pathlib.Path) -> pathlib.Path:
    '''Get the path to a directory, creating it if it doesn't exist.'''
    path.mkdir(parents=True, exist_ok=True)
    assert path.is_dir(), f'Directory {path} does not exist.'
    return path


class GameFiles:
    '''Access game-specific configuration files and directories.'''

    def __init__(self, name: str):
        '''Get files and directories for a game.

        :param name: Name of the game.
        :type name: str
        '''
        self.name = name
        self._cfg_data = None

    def server_dir(self) -> pathlib.Path:
        '''Get the path to the game server's runtime directory.'''
        path = paths.get('server-hot') / self.name
        return force_dir(path)

    def cfg_dir(self) -> pathlib.Path:
        '''Get the path to the game's configuration directory.'''
        path = paths.get('cfg') / self.name
        return force_dir(path)

    def backup_dir(self) -> pathlib.Path:
        '''Get the path to the game's backup directory.'''
        path = paths.get('backup') / self.name
        return force_dir(path)

    def shelf_dir(self) -> pathlib.Path:
        '''Get the path to the game's shelf directory.'''
        path = paths.get('shelf') / self.name
        return force_dir(path)

    def cfg_file(self) -> pathlib.Path:
        '''Get the path to the game's configuration file.'''
        path = self.cfg_dir() / f'{self.name}.json'
        return path

    def cfg_data(self) -> dict:
        '''Get the game's configuration data.'''
        data = self._cfg_data

        if data is None:
            cfg = self.cfg_file()

            with open(cfg, 'r', encoding='utf-8') as file:
                data = json.load(file)

        self._cfg_data = data
        return data

    def dockerfile(self) -> pathlib.Path:
        '''Get the path to the game server Dockerfile.'''
        dockerfile = self.cfg_dir() / f'{self.name}.dockerfile'
        return dockerfile

    def start_script(self) -> pathlib.Path:
        '''Get the path to the game server start script.'''
        script = self.cfg_dir() / 'start.sh'
        return script

    def download_script(self) -> pathlib.Path:
        '''Get the path to the game server download script.'''
        script = self.cfg_dir() / 'download.sh'
        return script

    def backup_script(self) -> pathlib.Path:
        '''Get the path to the game server backup script.'''
        script = self.cfg_dir() / 'backup.sh'
        return script

    def restore_script(self) -> pathlib.Path:
        '''Get the path to the game server restore script.'''
        script = self.cfg_dir() / 'restore.sh'
        return script
