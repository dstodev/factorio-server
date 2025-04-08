'''CLI entrypoint to the manage module.'''

from cli import cli

from manage import paths
from manage.game_files import GameFiles


def main():
    print('Hello, world!')
    args = cli.args()

    print(args)

    # client = docker.from_env()

    games = {}

    for path in paths.get('cfg').iterdir():
        if path.is_dir():
            name = path.name
            print(f'Found game: {name}')
            files = GameFiles(name)

            print(files.cfg_dir())
            print(files.backup_dir())
            print(files.shelf_dir())

            print(files.cfg_file())
            print(files.cfg_data())
            print(files.dockerfile())
            print(files.start_script())
            print(files.download_script())
            print(files.backup_script())
            print(files.restore_script())

            games[name] = {
                'files': files,
            }

    # print(client.containers.run('hello-world').decode('utf-8'))


if __name__ == '__main__':
    main()
