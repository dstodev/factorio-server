'''Paths to important project directories.'''

from pathlib import Path

this_dir = Path(__file__).resolve().parent
project_dir = this_dir.parent
repo_dir = project_dir.parent.parent


def get(which: str = '', root: Path = repo_dir) -> Path:
    '''Get the path to a project directory:

    :param which: Name of the directory to get:

        - `'src'`: Repo root
        - `'server-hot'`: Server directory; server may be actively running here
        - `'backup'`: Game backups
        - `'cfg'`: Game configurations
        - `'docker'`: Docker directory
        - `'rcon'`: RCON client
        - `'shelf'`: Game shelf directory

    :type which: str

    :param root: Root directory to to base the project directory on.
    :type root: pathlib.Path
    '''
    return {
        'src': root,
        'server-hot': root / 'server-files',
        'backup': root / 'backup',
        'cfg': root / 'cfg',
        'docker': root / 'docker',
        'rcon': root / 'rcon',
        'shelf': root / 'shelf',
    }[which]
