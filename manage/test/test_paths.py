from functools import partial

from manage import paths


def test_paths(tmp_path):
    get = partial(paths.get, root=tmp_path)

    assert get('src') == tmp_path

    assert get('backup') == tmp_path / 'backup'
    assert get('cfg') == tmp_path / 'cfg'
    assert get('docker') == tmp_path / 'docker'
    assert get('rcon') == tmp_path / 'rcon'
    assert get('shelf') == tmp_path / 'shelf'
