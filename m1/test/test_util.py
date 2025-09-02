'''Test miscellaneous utility functions from: manage.util'''

from manage.util import dir_has_files


def test_dir_has_files(tmp_path, tmp_file):
    assert not dir_has_files(tmp_path)
    tmp_file('test.txt', 'Hello, World!')
    assert dir_has_files(tmp_path)

    some_dir = tmp_path / 'some-dir'
    assert not dir_has_files(some_dir)
    some_dir.mkdir()
    assert not dir_has_files(some_dir)
    tmp_file('some-dir/test.txt', 'Hello, World!')
    assert dir_has_files(some_dir)
