'''CLI entrypoint to the manage module.'''

from cli import cli

from manage import game, paths, rcon
from manage.command import Backup, Download, Shelf, Start, Stop
from manage.docker import build_image
from manage.util import clean_docker


def main():
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

        case 'rcon':
            if args.send is not None:
                print(f'Sending RCON command: {args.send}')
                result = rcon.send(args.name, args.send)
                print(result)

            if args.say is not None:
                cfg = game.cfg_data(args.name)
                say = cfg['rcon']['say']
                say = say.replace('::', ' '.join(args.say))
                print(f'Sending RCON command: {say}')
                result = rcon.send(args.name, [say])
                print(result)

        case 'shelf':
            shelf = Shelf(args.name)
            shelf.execute()

        case _:
            print('Unknown command')


if __name__ == '__main__':
    main()
