'''Test shell utilities.'''

import pytest

from manage import shell


def test_run_file(tmp_file, uncap):
    file = tmp_file('script.sh',
                    '#!/bin/bash',
                    'echo "Hello,"',
                    'echo "World!" >&2',
                    mode=0o744)  # -rwxr--r--

    uncap(file)

    output = shell.run_file(file)

    assert output.stdout == 'Hello,\n'
    assert output.stderr == 'World!\n'


def test_run_file_error(tmp_file, uncap):
    file = tmp_file('script.sh',
                    '#!/bin/bash',
                    'echo "Hello,"',
                    'echo "World!" >&2',
                    'exit 1',
                    mode=0o744)  # -rwxr--r--

    uncap(file)

    try:
        shell.run_file(file)
    except shell.CalledProcessError as e:
        assert e.returncode == 1
        assert e.stdout == 'Hello,\n'
        assert e.stderr == 'World!\n'


def test_run_file_args(tmp_file, uncap):
    file = tmp_file('script.sh',
                    '#!/bin/bash',
                    'echo "$1"',
                    'echo "$2"',
                    mode=0o744)  # -rwxr--r--

    uncap(file)

    output = shell.run_file(file, 'Hello,', 'World!')

    assert output.stdout == 'Hello,\nWorld!\n'
    assert output.stderr == ''


def test_run_file_no_execute_permission(tmp_file):
    file = tmp_file('text-file.txt',
                    'Hello, World!',
                    mode=0o644)  # -rw-r--r--

    with pytest.raises(PermissionError):
        shell.run_file(file)


def test_run_file_not_found(tmp_path):
    file = tmp_path / 'non-existent-script.sh'

    with pytest.raises(FileNotFoundError):
        shell.run_file(file)
