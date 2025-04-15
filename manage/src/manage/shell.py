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


def run_file(script: pathlib.Path, *args: str) -> Result:
    '''Run a script with provided arguments.

    :param script: The script to run.
    :type script: str
    :param args: Arguments to pass to the script.
    :type args: str
    '''
    result = subprocess.run([script, *args], check=False, capture_output=True, text=True)
    return Result(result.returncode, result.stdout, result.stderr)
