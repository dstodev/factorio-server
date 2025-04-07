import json
import pathlib

from manage import paths


def force_dir(dir: pathlib.Path):
    dir.mkdir(parents=True, exist_ok=True)
    assert dir.is_dir(), f'Directory {dir} does not exist.'
    return dir


def assert_file(file: pathlib.Path):
    assert file.is_file(), f'File {file} does not exist.'
    return file


class Game:
    def __init__(self, name: str):
        self.name = name
        self._cfg_data = None

    def cfg_dir(self):
        dir = paths.get('cfg') / self.name
        return force_dir(dir)

    def backup_dir(self):
        dir = paths.get('backup') / self.name
        return force_dir(dir)

    def shelf_dir(self):
        dir = paths.get('shelf') / self.name
        return force_dir(dir)

    def cfg_file(self):
        cfg = self.cfg_dir() / f'{self.name}.json'
        return assert_file(cfg)

    def cfg_data(self):
        name = self.name
        data = self._cfg_data

        if data is None:
            try:
                cfg = self.cfg_file()

                with open(cfg, 'r') as file:
                    data = json.load(file)

            except json.JSONDecodeError:
                raise ValueError(f'{name} cfg {cfg} does not contain valid JSON.')

            except Exception as e:
                raise e

            if not isinstance(data, dict):
                raise ValueError(f'Game cfg {cfg} does not contain valid JSON.')

        self._cfg_data = data
        return data

    def dockerfile(self):
        dockerfile = self.cfg_dir() / f'{self.name}.dockerfile'
        return assert_file(dockerfile)

    def start_script(self):
        script = self.cfg_dir() / 'start.sh'
        return assert_file(script)

    def download_script(self):
        script = self.cfg_dir() / 'download.sh'
        return assert_file(script)

    def backup_script(self):
        script = self.cfg_dir() / 'backup.sh'
        return assert_file(script)

    def restore_script(self):
        script = self.cfg_dir() / 'restore.sh'
        return assert_file(script)
