'''Tests for build_args function'''

import json
from functools import partial

from manage import paths
from manage.command.download import Download


def test_download(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    dockerfile = tmp_file('cfg/test-game/server.dockerfile',
                          'FROM alpine:latest',
                          'ARG user_id',
                          'ARG user_name',
                          'ARG group_id',
                          'ARG group_name',
                          'RUN addgroup -g $group_id $group_name \\',
                          '  && adduser -u $user_id -D -G $group_name $user_name',
                          'USER $user_name')

    script = tmp_file('cfg/test-game/download.sh',
                      '#!/bin/sh',
                      'server_dir="$1"',
                      'touch "$server_dir/some-server-file"',
                      mode=0o744)

    expected_uid = 30121
    expected_gid = 30122

    server_json = tmp_file('cfg/test-game/server.json',
                           json.dumps({
                               'user': {
                                   'name': f'server-user:{expected_uid}',
                                   'group': f'server-group:{expected_gid}'
                               }
                           }, indent=2))

    uncap(dockerfile)
    uncap(script)
    uncap(server_json)

    expected_server_dir = tmp_path / 'server-files/test-game'

    uncap(expected_server_dir)

    name = 'test-game'

    download = Download(name)

    download.execute()

    assert expected_server_dir.is_dir(), 'Server directory was not created'

    assert expected_server_dir.stat().st_uid == expected_uid, f'Server directory user id is not {expected_uid}'
    assert expected_server_dir.stat().st_gid == expected_gid, f'Server directory group id is not {expected_gid}'

    expected_server_file = expected_server_dir / 'some-server-file'
    assert expected_server_file.is_file(), 'Server file was not created'
    assert expected_server_file.stat().st_uid == expected_uid, f'Server file user id is not {expected_uid}'
    assert expected_server_file.stat().st_gid == expected_gid, f'Server file group id is not {expected_gid}'
