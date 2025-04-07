from pathlib import Path

this_dir = Path(__file__).resolve().parent
project_dir = this_dir.parent
repo_dir = project_dir.parent


def get(which: str = '', root: Path = repo_dir) -> Path:
    return {
        'src': root,

        'backup': root / 'backup',
        'cfg': root / 'cfg',
        'docker': root / 'docker',
        'rcon': root / 'rcon',
        'shelf': root / 'shelf',
    }[which]
