'''Test custom test fixtures.'''


def test_tree(tmp_path, uncap, tree):
    uncap(tmp_path)

    structure = {
        'dir1': {
            'file1.txt': 'Hello, World!',
            'file2.txt': [
                'Line 1',
                'Line 2'
            ],
            'file3.txt': None,
            'file4.txt': '',
            'subdir': {
                'file.txt': 'a',
            }
        },
        'dir2': {
            'file1.txt': 'b',
            'file2.txt': 'c'
        },
        'dir3/subdir': {
            'file.txt': 'd'
        },
        'dir4/subdir/file.txt': 'e'
    }
    tree(structure)

    assert (tmp_path / 'dir1' / 'file1.txt').read_text() == 'Hello, World!\n'
    assert (tmp_path / 'dir1' / 'file2.txt').read_text() == 'Line 1\nLine 2\n'
    assert (tmp_path / 'dir1' / 'file3.txt').read_text() == ''
    assert (tmp_path / 'dir1' / 'file4.txt').read_text() == ''
    assert (tmp_path / 'dir1' / 'subdir' / 'file.txt').read_text() == 'a\n'

    assert (tmp_path / 'dir2' / 'file1.txt').read_text() == 'b\n'
    assert (tmp_path / 'dir2' / 'file2.txt').read_text() == 'c\n'

    assert (tmp_path / 'dir3' / 'subdir' / 'file.txt').read_text() == 'd\n'
    assert (tmp_path / 'dir4' / 'subdir' / 'file.txt').read_text() == 'e\n'


def test_tmp_file(tmp_path, tmp_file, uncap):
    uncap(tmp_path)

    file = tmp_file('test1.txt')
    assert file == tmp_path / 'test1.txt'
    assert file.read_text() == ''

    file = tmp_file('test2.txt', 'Line 1')
    assert file == tmp_path / 'test2.txt'
    assert file.read_text() == 'Line 1\n'
    assert file.stat().st_mode & 0o777 == 0o644

    file = tmp_file('test3.txt', 'Line 1', 'Line 2', mode=0o600)
    assert file == tmp_path / 'test3.txt'
    assert file.read_text() == 'Line 1\nLine 2\n'
    assert file.stat().st_mode & 0o777 == 0o600

    file = tmp_file('subdir/subdir/test.txt')
    assert file == tmp_path / 'subdir/subdir' / 'test.txt'
    assert file.exists()

    file = tmp_file('test4.txt', mode=0o600)
    assert file == tmp_path / 'test4.txt'
    assert file.stat().st_mode & 0o777 == 0o600
