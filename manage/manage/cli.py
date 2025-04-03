import docker

import manage.parse as parse


def main():
    print("Hello, world!")
    args = parse.args()

    print(args)

    client = docker.from_env()

    print(client.containers.run('hello-world').decode('utf-8'))


if __name__ == '__main__':
    main()
