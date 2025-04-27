'''Command to download the server files.'''

import os
import stat

from manage import game_files
from manage.docker.build_args import build_args
from manage.docker.run_container import Bind, RunContainer
from manage.shell import Result


class Download:
    '''Run the download script for the game.

    The game's download script runs in and ephemeral container based on the
    server image, and assumes it is running as a non-root server user. The
    script receives one argument: the path to the server root directory to
    populate.

    The script's responsiblity is to prepare the server root directory with all
    files needed to run the server, and to set up any initial details like
    configuration files, mods, permissions, etc. (Since the container runs as
    the server user, permissions should already be correct.)

    The script need not be idempotent, meaning it does not need to care about
    what to do when downloading to an already-populated directory, because this
    class will only call it when populating a fresh server directory.
    '''

    def __init__(self, game: str) -> None:
        '''Initialize the download command.

        :param game: The game to download.
        :type game: str
        '''
        self.game = game

        self.server_dir = game_files.server_dir(game)

        binds = [
            Bind(host=self.server_dir.parent,
                 guest='/parent',
                 writeable=True),
            Bind(host=game_files.download_script(game),
                 guest='/download.sh',
                 writeable=False),
        ]

        self.container = RunContainer(name=game,
                                      dockerfile_path=game_files.dockerfile(game),
                                      build_args=build_args(game),
                                      binds=binds)

    def execute(self) -> None:
        '''Run the download script.'''

        if not self.server_dir.exists():
            # Ensure the shared server root directory exists
            parent_dir = self.server_dir.parent
            old_umask = os.umask(0o000)  # set umask to 0o000 to allow full permissions
            # create with full permissions (needs o+w so server user can write)
            parent_dir.mkdir(parents=True, exist_ok=True, mode=0o777)
            os.umask(old_umask)  # reset umask to previous value
            os.chmod(parent_dir, os.stat(parent_dir).st_mode | stat.S_ISVTX)  # set sticky bit

            # root must must align with guest value for server_dir_parent bind mount in __init__
            guest_server_dir = f'/parent/{self.game}'

            result = self.container.run(
                entrypoint=['/bin/sh', '-c'],
                command=[' && '.join((
                    # create the server directory in the container so it is owned by the server user
                    f'mkdir {guest_server_dir}',
                    'cp /download.sh /tmp/download.sh',
                    f'/tmp/download.sh {guest_server_dir}',
                ))])

            assert isinstance(result, Result)

            if result.exit_status != 0:
                msg = ''.join((
                    '\n== Error! ======================================================================',
                    f'\nDownload script failed with exit code: {result.exit_status}',
                    '\n-- stdout: ---------------------------------------------------------------------',
                    f'\n{result.stdout.strip()}' if result.stdout else '',
                    '\n-- stderr: ---------------------------------------------------------------------',
                    f'\n{result.stderr.strip()}' if result.stderr else '',
                    '\n================================================================================',
                ))
                raise RuntimeError(msg)
