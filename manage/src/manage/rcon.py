'''RCON helpers.'''


import os
import random
import string

from manage.docker.container import ExecResult, GameContainer
from manage.game import cfg_data, cfg_dir


class RconError(Exception):
    pass


def send(name: str, command: list[str]) -> ExecResult | None:
    '''Send an RCON command to a game server.

    :param name: The game to send to.
    :type name: str

    :param command: The command to send.
    :type command: str

    :return: The result of the command.

    :rtype: ExecResult
    :raises RconError: The RCON command did not send successfully.
    '''
    # TODO: How to better reconcile the container name being different from the game name?

    cfg = cfg_data(name)

    try:
        cfg_port = cfg['port']
        port = cfg_port['rcon']
    except KeyError as e:
        raise RconError(f'No RCON port found in configuration for game: {name}') from e

    container = GameContainer(f'{name}-server', None)  # Only reattach
    hoststr = f'127.0.0.1:{port}'
    pw = password(name)

    try:
        result = container.execute(['rcon', hoststr, *command], send_stdin=pw)
    except RuntimeError as e:
        raise RconError('No target container!') from e

    if result is not None:
        code = result.exit_status

        if code != 0:
            raise RconError(f'RCON client failure ({code}): {result.output}')

    return result


def password(name: str, length: int = 128, new: bool = False) -> str:
    '''Get the RCON password for the game server.

    if new is True, a password is generated with the provided length, saving it
    to the secret file in the game's configuration directory before returning
    it.
    '''
    cfg = cfg_dir(name)
    secret_file = cfg / 'secret'

    try:
        if new:
            raise FileNotFoundError

        with open(secret_file, 'r', encoding='utf-8') as file:
            pw = file.read().strip()

    except FileNotFoundError:
        char_pool = string.ascii_letters + string.digits
        pw = ''.join(random.choices(char_pool, k=length))

        secret_file.parent.mkdir(parents=True, exist_ok=True)

        secret_fd = os.open(secret_file, os.O_CREAT | os.O_WRONLY | os.O_TRUNC, mode=0o600)

        with os.fdopen(secret_fd, 'w', encoding='utf-8') as file:
            file.write(f'{pw}\n')

    return pw
