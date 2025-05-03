'''Test the RCON Command.'''


import json
import shutil
from functools import partial
from test.util_docker import clean_docker

import pytest

from manage import game, paths
from manage.command import Rcon
from manage.docker import GameContainer, build_image


def test_rcon_command(mocker, tmp_path, tmp_file, uncap):
    shutil.copytree(paths.get('rcon'), tmp_path / 'rcon')
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-rcon-command'
    rcon_image_name = 'rcon-test-rcon-command'

    dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                          f'FROM {rcon_image_name}:latest')

    expected_uid = 30120
    expected_gid = 30121

    server_json = tmp_file(f'cfg/{name}/server.json',
                           json.dumps({
                               'user': {
                                   'name': f'server-user:{expected_uid}',
                                   'group': f'server-group:{expected_gid}'
                               }
                           }, indent=2))

    uncap(dockerfile)
    uncap(server_json)

    try:
        build_image(paths.get('rcon') / 'Dockerfile', rcon_image_name, game.build_args(name))

        image, _logs = game.docker_image(name)

        container = GameContainer(name, image)

        container.start(command=['tail', '-f', '/dev/null'])

        rcon = Rcon(name, ['test'])
        rcon.execute()
        result = rcon.last_result

        assert result is not None
        assert result.exit_code == 0
        assert result.output == 'All tests passed!\n'

        result = container.execute(['/bin/sh', '-c', 'stat -c "%u:%g" "$(which rcon)"'])

        assert result is not None
        assert result.exit_code == 0
        assert result.output == '0:0\n'

        result = container.execute(['/bin/sh', '-c', 'echo "$(id -u):$(id -g)"'])

        assert result is not None
        assert result.exit_code == 0
        assert result.output == f'{expected_uid}:{expected_gid}\n'

        assert container.container is not None
        container.container.stop()
        container.wait()

        assert container.container is None

    finally:
        clean_docker(name)
        # clean_docker(rcon_image_name)


def test_rcon_command_no_container(mocker, tmp_path):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'

    rcon = Rcon(name, ['test'])

    with pytest.raises(RuntimeError, match='Container does not exist'):
        rcon.execute()
