'''Test the RCON Command.'''


import json
import shutil
from functools import partial
from test.util_docker import clean_docker

from manage import game, paths
from manage.command import Rcon
from manage.docker import ServerContainer, build_image


def test_rcon_command(mocker, tmp_path, tmp_file, uncap):
    shutil.copytree(paths.get('rcon'), tmp_path / 'rcon')
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game-rcon-command'

    try:
        dockerfile = tmp_file(f'cfg/{name}/server.dockerfile',
                              'FROM rcon:latest')

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

        build_image(paths.get('rcon') / 'Dockerfile', 'rcon', game.build_args(name))

        container = ServerContainer(name,
                                    game.dockerfile(name),
                                    game.build_args(name))

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

        container.execute(['pkill', 'tail'])
        container.wait()

        assert container.container is None

    finally:
        clean_docker(name)
