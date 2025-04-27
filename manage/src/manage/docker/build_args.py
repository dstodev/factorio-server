'''Common arguments for building game dockerfiles.

Dockerfiles commonly include the lines:
ARG game_port
ARG rcon_port
ARG user_id
ARG user_name
ARG group_id
ARG group_name
'''

from manage import game_files


def build_args(game: str) -> dict[str, str]:
    '''Return common build arguments for game dockerfiles.'''

    args = {}

    cfg = game_files.cfg_data(game)

    try:
        fields = {}
        ports = cfg['port']
        fields['game_port'] = str(ports['game'])
        fields['rcon_port'] = str(ports['rcon'])
        args.update(fields)
    except KeyError:
        pass

    try:
        fields = {}
        user = cfg['user']
        fields['user_name'], fields['user_id'] = user['name'].split(':')
        fields['group_name'], fields['group_id'] = user['group'].split(':')
        args.update(fields)
    except KeyError:
        pass

    return args
