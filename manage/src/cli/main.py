'''CLI entrypoint to the manage module.'''

from cli import cli

from manage import game, paths
from manage.command import Backup, Download, Start, Stop
from manage.docker import build_image
from manage.util import clean_docker


def main():
    print('Hello, world!')
    args = cli.args()

    print(args)

    # client = docker.from_env()

    for path in paths.get('cfg').iterdir():
        if path.is_dir():
            name = path.name
            print(f'Found game: {name}')

    # Update rcon image with game's build args
    build_image(paths.get('rcon') / 'Dockerfile', 'rcon', game.build_args(args.name))

    match args.command:
        case 'download':
            try:
                download = Download(args.name)
                download.execute()
            finally:
                clean_docker(f'{args.name}-download')

        case 'start':
            start = Start(args.name)
            start.execute()

        case 'stop':
            stop = Stop(args.name)
            stop.execute()

        case 'backup':
            backup = Backup(args.name)
            backup.execute()

        case _:
            print('Unknown command')


if __name__ == '__main__':
    main()
