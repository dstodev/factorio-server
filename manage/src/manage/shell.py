'''Shell interface tools for e.g. running scripts.'''

import pathlib
import subprocess
from subprocess import CalledProcessError
from typing import NamedTuple

assert CalledProcessError is not None  # suppress unused import


class Result(NamedTuple):
    '''Collected output from running a program.'''
    exit_status: int
    stdout: str
    stderr: str


def run_file(path: pathlib.Path, *args: str) -> Result:
    '''Run a file with provided arguments.

    :param path: The file to run.
    :type script: pathlib.Path
    :param args: Arguments to pass to the file.
    :type args: str
    '''
    result = subprocess.run([path, *args], check=False, capture_output=True, text=True)
    return Result(result.returncode, result.stdout, result.stderr)
