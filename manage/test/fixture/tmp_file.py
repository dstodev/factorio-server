'''Pytest fixture to create a temporary file with provided content.'''

import os
import pathlib

import pytest

SEP = os.linesep


@pytest.fixture
def tmp_file(tmp_path):
    '''Create a temporary file with provided content.'''

    # 0o644 = -rw-r--r--
    def _create_file(name: str, *content: str, mode: int = 0o644) -> pathlib.Path:
        '''Create a temporary file with the given name and content.
        Each string passed as content represents the next line of file content.
        '''
        file = tmp_path / name
        file.parent.mkdir(parents=True, exist_ok=True)
        file.touch()
        if content:
            file.write_text(SEP.join(content) + SEP)
            cur_mode_masked = file.stat().st_mode & ~0o777
            file.chmod(cur_mode_masked | mode)
        return file

    return _create_file
