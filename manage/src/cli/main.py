'''CLI entrypoint to the manage module.'''

from argparse import Namespace

from cli import cli

from manage import game, paths, rcon
from manage.command import Backup, Download, Shelf, Start, Stop
from manage.docker import build_image


def main():
    args: Namespace = cli.args()

    if args.verbose and len(args.__dict__) > 0:
        print('Options:')
        for key, value in args.__dict__.items():
            if value:
                print(f'  {key}: {value}')

    if args.list:
        for path in paths.get('cfg').iterdir():
            if path.is_dir():
                name = path.name
                print(f'{name}')
        return

    if args.name is None:
        print('No game name specified. Use cli -l to list available games.')
        return

    # Update rcon image with game's build args
    build_image(paths.get('rcon') / 'Dockerfile', 'rcon', game.build_args(args.name))

    action = None

    match args.command:
        case 'download':
            action = Download(args.name)

        case 'start':
            action = Start(args.name)

        case 'stop':
            action = Stop(args.name)

        case 'backup':
            action = Backup(args.name)

        case 'shelf':
            action = Shelf(args.name)

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

        case _:
            print('Unknown command')

    if action is not None:
        if args.verbose:
            print(f'Executing {action.__class__.__name__} for game: {args.name}')
        action.execute()


if __name__ == '__main__':
    main()
