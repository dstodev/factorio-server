import json

from manage import BACKUP_DIR, CFG_DIR, SHELF_DIR


class Game:
    def __init__(self, name: str):
        self.name = name
        self._config_data = None

    def config_path(self):
        name = self.name
        path = CFG_DIR / name / f'{name}.json'
        assert path.is_file(), f'Config file {path} does not exist.'
        return path

    def config_data(self):
        name = self.name
        data = self._config_data
        if data is None:
            try:
                path = self.config_path()
                with open(path, 'r') as file:
                    data = json.load(file)
            except json.JSONDecodeError:
                raise ValueError(f'{name} config {path} does not contain valid JSON.')
            except Exception as e:
                raise e

            if not isinstance(data, dict):
                raise ValueError(f'Game config {path} does not contain valid JSON.')

        self._config_data = data
        return data

    def start_script(self):
        path = CFG_DIR / self.name / 'start.sh'
        assert path.is_file(), f'Start script {path} does not exist.'
        return path

    def download_script(self):
        path = CFG_DIR / self.name / 'download.sh'
        assert path.is_file(), f'Download script {path} does not exist.'
        return path

    def backup_script(self):
        path = CFG_DIR / self.name / 'backup.sh'
        assert path.is_file(), f'Backup script {path} does not exist.'
        return path

    def restore_script(self):
        path = CFG_DIR / self.name / 'restore.sh'
        assert path.is_file(), f'Restore script {path} does not exist.'
        return path

    def backup_dir(self):
        path = BACKUP_DIR / self.name
        path.mkdir(parents=True, exist_ok=True)
        assert path.is_dir(), f'Backup path {path} does not exist.'
        return path

    def shelf_dir(self):
        path = SHELF_DIR / self.name
        path.mkdir(parents=True, exist_ok=True)
        assert path.is_dir(), f'Shelf {path} does not exist.'
        return path
