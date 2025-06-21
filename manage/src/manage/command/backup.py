'''Command to back up a game server.'''

import stat

from manage import game
from manage.command.rcon import Rcon
from manage.command.schedule import Schedule
from manage.docker import Bind, GameContainer
from manage.rcon import RconError
from manage.shell import Result
from manage.util import dir_has_files, timestamp


class Backup:
    '''Back up a game server.

    Tries to save with RCON before backing up.
    Does not require the server to be running.
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

        self.last_result = None

        self.try_save_pre = Schedule()
        self.try_save_post = Schedule()

        try:
            cfg_rcon = game.cfg_data(name)['rcon']

            for key in ('save-pre', 'save'):
                try:
                    self.try_save_pre.add_action(Rcon(name, [cfg_rcon[key]]))
                except KeyError:
                    pass

            for key in ('save-post',):
                try:
                    self.try_save_post.add_action(Rcon(name, [cfg_rcon[key]]))
                except KeyError:
                    pass
        except KeyError:
            pass

    def execute(self) -> None:
        '''Run the backup script.'''
        if dir_has_files(self.server_dir):
            time = timestamp()

            guest_server_dir = '/game/hot'
            guest_backup_dir = f'/backup/{time}'

            # TODO: Test timestamp directory already exists (time collision)
            backup_dir = self.backup_dir / time
            backup_dir.mkdir(parents=True, exist_ok=True)

            try:
                self.try_save_pre.execute()
            except RconError:
                pass

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

                result = self.container.wait(timeout=600)
                self.last_result = result

            finally:
                backup_dir.chmod(restore_mode)

                try:
                    self.try_save_post.execute()
                except RconError:
                    pass

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
