'''CLI argument parser.'''

import argparse


def args():
    '''Get user-specified command-line options.'''
    parser = argparse.ArgumentParser(prog='manage',
                                     description='Server management suite',
                                     epilog='')
    subparsers = parser.add_subparsers(dest='command',
                                       help='which tool to use')

    parser_new = subparsers.add_parser('new', help='create a new server')
    parser_new.add_argument('name', help='name of the game')

    parser_start = subparsers.add_parser('start', help='start a server')
    parser_start.add_argument('name', help='name of the game')

    parser_stop = subparsers.add_parser('stop', help='stop a server')
    parser_stop.add_argument('name', help='name of the game')
    parser_stop.add_argument('--force', action='store_true',
                             help='force stop the server')

    parser_stop.add_argument('--timeout', type=int, default=10,
                             help='time to wait after notifying players before stopping the server')

    parser_backup = subparsers.add_parser('backup', help='backup a server')
    parser_backup.add_argument('name', help='name of the game')
    # Force backup: Normally stops if cannot RCON save first to avoid saving an offline server
    #               Force to backup even if the server is offline/not responding to RCON
    parser_backup.add_argument('--force', action='store_true',
                               help='force backup the server')

    parser_shelf = subparsers.add_parser('shelf', help='manage a server shelf')
    parser_shelf.add_argument('name', help='name of the game')
    parser_shelf.add_argument('--push', action='store_true',
                              help='push the server to the shelf')
    parser_shelf.add_argument('--restore', action='store_true',
                              help='restore the server from the shelf')
    parser_shelf.add_argument('--latest', action='store_true',
                              help='print the latest server path on the shelf')
    parser_shelf.add_argument('--list', action='store_true',
                              help='list all server instances on the shelf')

    parser_download = subparsers.add_parser('download', help='download a server')
    parser_download.add_argument('name', help='name of the game')

    parser_rcon = subparsers.add_parser('rcon', help='send a command to the server')
    parser_rcon.add_argument('name', help='name of the game')
    parser_rcon.add_argument('--send', nargs='*',  # 0 args is useful as ping
                             help='send a command to the server')
    parser_rcon.add_argument('--save', action='store_true',
                             help='save the server')
    parser_rcon.add_argument('--stop', action='store_true',
                             help='stop the server')
    parser_rcon.add_argument('--say', nargs='+',
                             help='say something on the server')

    return parser.parse_args()
