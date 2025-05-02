from functools import partial

from manage import paths
from manage.command import Shelf


def test_shelf(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'
    server_dir = tmp_path / f'server-files/{name}'

    some_file = tmp_file(f'server-files/{name}/hot/some-file.txt',
                         'Hello!')

    uncap(some_file)

    assert server_dir.exists()

    shelf = Shelf(name)
    shelf.execute()

    expected_shelf_dir = tmp_path / f'shelf/{name}/1'
    expected_file = expected_shelf_dir / 'hot/some-file.txt'

    assert expected_shelf_dir.is_dir()
    assert expected_file.is_file()
    assert expected_file.read_text() == 'Hello!\n'
    assert not some_file.is_file()  # File was moved, not copied
    assert not server_dir.exists()  # Server dir was moved
    assert server_dir.parent.exists()


def test_shelf_twice(mocker, tmp_path, tmp_file, uncap):
    mocker.patch('manage.paths.get', side_effect=partial(paths.get, root=tmp_path))

    name = 'test-game'
    server_dir = tmp_path / f'server-files/{name}'

    count = 0

    def do():
        nonlocal count
        count += 1

        some_file = tmp_file(f'server-files/{name}/hot/some-file.txt',
                             f'Count: {count}')

        uncap(some_file)

        assert server_dir.exists()

        shelf = Shelf(name)
        shelf.execute()

        expected_shelf_dir = tmp_path / f'shelf/{name}/{count}'
        expected_file = expected_shelf_dir / 'hot/some-file.txt'

        uncap(expected_file)

        assert expected_shelf_dir.is_dir()
        assert expected_file.is_file()
        assert expected_file.read_text() == f'Count: {count}\n'
        assert not some_file.is_file()
        assert not server_dir.exists()
        assert server_dir.parent.exists()

    do()
    do()
