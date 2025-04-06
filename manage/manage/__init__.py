from pathlib import Path

this_dir = Path(__file__).resolve().parent


SRC_DIR = this_dir.parent.parent

BACKUP_DIR = SRC_DIR / 'backup'
CFG_DIR = SRC_DIR / 'cfg'
DOCKER_DIR = SRC_DIR / 'docker'
RCON_DIR = SRC_DIR / 'rcon'
SHELF_DIR = SRC_DIR / 'shelf'

assert BACKUP_DIR.is_dir(), f'Backup directory {BACKUP_DIR} does not exist.'
assert CFG_DIR.is_dir(), f'Configuration directory {CFG_DIR} does not exist.'
assert DOCKER_DIR.is_dir(), f'Docker directory {DOCKER_DIR} does not exist.'
assert RCON_DIR.is_dir(), f'RCON directory {RCON_DIR} does not exist.'
assert SHELF_DIR.is_dir(), f'Shelf directory {SHELF_DIR} does not exist.'
