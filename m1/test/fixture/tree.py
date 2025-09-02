'''Set up file structures for testing.'''

import pathlib
from collections.abc import Iterable

import pytest


@pytest.fixture
def tree(tmp_path):
    '''Create a temporary file structure.'''

    def _create_tree(structure: dict, root: pathlib.Path = tmp_path) -> None:
        '''Create a file structure based on the provided dictionary.

        - Keys represent directories or files, and values represent their content.
        - Dictionary values represent subdirectories.
        - Strings or arrays of strings represent file contents. Adds newlines.

        example:

            tree({
                'dir1': {
                    'file1.txt': 'Hello, World!',
                    'file2.txt': [
                        'Line 1',
                        'Line 2'
                    ],
                    'subdir': {
                        'file.txt': None
                    }
                },
                'dir2': {
                    'file.txt': None
                }
            })

            creates the following structure & content:

            /tmp/path
            ├── dir1
            │   ├── file1.txt:
            │   │       Hello, World!
            │   ├── file2.txt:
            │   │       Line 1
            │   │       Line 2
            │   └── subdir
            │       └── file.txt
            └── dir2
                └── file.txt
        '''
        for key, value in structure.items():
            path = root / key
            path.parent.mkdir(parents=True, exist_ok=True)

            if isinstance(value, dict):
                _create_tree(value, path)
            elif value is None or value == '':
                path.touch()
            else:
                if isinstance(value, Iterable) \
                        and not isinstance(value, (str, bytes)):
                    value = '\n'.join(value)
                path.write_text(str(value) + '\n')

    return _create_tree
