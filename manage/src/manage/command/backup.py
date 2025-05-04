'''Command to back up a game server.'''

import datetime
import stat

from manage import game
from manage.docker import Bind, GameContainer
from manage.shell import Result


class Backup:
    '''Back up a game server.

    Does not issue a save command or require the server to be running.
    '''

    def __init__(self, name: str) -> None:
        '''Initialize the backup command.

        :param name: The game to back up.
        :type name: str
        '''
        self.name = name

        # server_dir looks like: <repo>/server-files/<game-name>/hot
        self.server_dir = game.server_dir(name)

        # backup_dir looks like: <repo>/backup/<game-name>
        self.backup_dir = game.backup_dir(name)

        binds = [
            Bind(host=self.server_dir.parent, guest='/game', writeable=False),
            Bind(host=self.backup_dir, guest='/backup', writeable=True),
            Bind(host=game.backup_script(name), guest='/backup.sh', writeable=False),
        ]

        image, _logs = game.docker_image(name)

        self.container = GameContainer(f'{name}-backup', image, binds)

    def execute(self) -> None:
        '''Run the backup script.'''
        timestamp = datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%d_%H-%M-%SZ')

        guest_server_dir = '/game/hot'
        guest_backup_dir = f'/backup/{timestamp}'

        # TODO: Test timestamp directory already exists
        backup_dir = self.backup_dir / timestamp
        backup_dir.mkdir(parents=True, exist_ok=True)

        command = ' && '.join([
            'cp /backup.sh /tmp/backup.sh',
            f'/tmp/backup.sh {guest_server_dir} {guest_backup_dir}'
        ])

        restore_mode = self.backup_dir.stat().st_mode

        try:
            # Temporarily set full permissions (o+w so server user can write) & set sticky bit
            backup_dir.chmod(restore_mode | 0o777 | stat.S_ISVTX)

            self.container.start(entrypoint=['/bin/sh', '-c'],
                                 command=[command])

            result = self.container.wait()
        finally:
            backup_dir.chmod(restore_mode)

        assert isinstance(result, Result)

        if result.exit_status != 0:
            msg = ''.join((
                '\n== Error! ======================================================================',
                f'\nBackup script failed with exit code: {result.exit_status}',
                '\n-- stdout: ---------------------------------------------------------------------',
                f'\n{result.stdout.strip()}' if result.stdout else '',
                '\n-- stderr: ---------------------------------------------------------------------',
                f'\n{result.stderr.strip()}' if result.stderr else '',
                '\n================================================================================'
            ))
            raise RuntimeError(msg)
