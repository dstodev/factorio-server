'''Shell interface tools for e.g. running scripts.'''

import pathlib
import subprocess
from subprocess import CalledProcessError
from typing import NamedTuple

assert CalledProcessError is not None  # suppress unused import


class Output(NamedTuple):
    '''Collected output from running a program.'''
    stdout: str
    stderr: str


def run_file(script: pathlib.Path, *args: str) -> Output:
    '''Run a script with provided arguments.

    :param script: The script to run.
    :type script: str
    :param args: Arguments to pass to the script.
    :type args: str
    '''

    result = subprocess.run([script, *args], check=True, capture_output=True, text=True)

    return Output(result.stdout, result.stderr)
