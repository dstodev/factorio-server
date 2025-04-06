# import docker

from pathlib import Path

from manage import CFG_DIR, parse
from manage.game import Game


def main():
    print("Hello, world!")
    args = parse.args()

    print(args)

    # client = docker.from_env()

    games = {}

    for dir in Path(CFG_DIR).iterdir():
        if dir.is_dir():
            name = dir.name
            print(f'Found game: {name}')
            game = Game(name)
            print(game.config_path())
            print(game.config_data())
            print(game.start_script())
            print(game.download_script())
            print(game.backup_script())
            print(game.restore_script())
            print(game.backup_dir())
            print(game.shelf_dir())
            games[name] = game

    # print(client.containers.run('hello-world').decode('utf-8'))


if __name__ == '__main__':
    main()
