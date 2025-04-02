import sys

import docker


def main():
    print("Hello, world!")
    print(sys.argv)

    client = docker.from_env()
    print(client.info())
