'''Command to download the server files.'''

import stat

from manage import game
from manage.docker import Bind, GameContainer
from manage.shell import Result
from manage.util import clean_docker


class Download:
    '''Run the download script for the game.

    The game's download script runs in an ephemeral container based on the
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

    def __init__(self, name: str) -> None:
        '''Initialize the Download command.

        :param name: The game to download.
        :type name: str
        '''
        self.server_dir = game.server_dir(name)

        image, _logs = game.docker_image(name)

        binds = [
            Bind(host=self.server_dir.parent, guest='/game', writeable=True),
            Bind(host=game.download_script(name), guest='/download.sh', writeable=False),
        ]

        self.container = GameContainer(f'{name}-download', image, binds=binds)

    def execute(self) -> None:
        '''Run the download script.'''

        if not self.server_dir.exists():
            # Ensure the shared server root directory exists
            parent_dir = self.server_dir.parent
            parent_dir.mkdir(parents=True, exist_ok=True)

            # /game aligns with guest value for server_dir.parent bind mount in __init__
            guest_server_dir = '/game/hot'

            # Create the server directory in the container so it is
            # owned by the server user.
            command = ' && '.join([
                'umask 0002',
                f'mkdir {guest_server_dir}',
                'cp /download.sh /tmp/download.sh',
                f'/tmp/download.sh {guest_server_dir}'
            ])

            restore_mode = parent_dir.stat().st_mode

            try:
                # Temporarily set full permissions (o+w so server user can write) & set sticky bit
                parent_dir.chmod(restore_mode | 0o777 | stat.S_ISVTX)

                self.container.start(entrypoint=['/bin/sh', '-c'],
                                     command=[command])

                result = self.container.wait(timeout=300)
            finally:
                parent_dir.chmod(restore_mode)
                clean_docker(self.container.name)

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
