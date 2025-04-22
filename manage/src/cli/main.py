'''CLI entrypoint to the manage module.'''

from cli import cli

from manage import paths


def main():
    print('Hello, world!')
    args = cli.args()

    print(args)

    # client = docker.from_env()

    for path in paths.get('cfg').iterdir():
        if path.is_dir():
            name = path.name
            print(f'Found game: {name}')


if __name__ == '__main__':
    main()
