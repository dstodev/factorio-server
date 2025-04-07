# import docker

from pathlib import Path

from cli import cli

from manage import paths
from manage.game import Game


def main():
    print('Hello, world!')
    args = cli.args()

    print(args)

    # client = docker.from_env()

    games = {}

    for dir in Path(paths.get('cfg')).iterdir():
        if dir.is_dir():
            name = dir.name
            print(f'Found game: {name}')
            game = Game(name)

            print(game.cfg_dir())
            print(game.backup_dir())
            print(game.shelf_dir())

            print(game.cfg_file())
            print(game.cfg_data())
            print(game.dockerfile())
            print(game.start_script())
            print(game.download_script())
            print(game.backup_script())
            print(game.restore_script())

            games[name] = game

    # print(client.containers.run('hello-world').decode('utf-8'))


if __name__ == '__main__':
    main()
