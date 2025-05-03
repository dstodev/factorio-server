'''Test running scripts from a Command-conforming class.'''

from manage.command import Script


def test_script(tmp_file):
    script = tmp_file('my_script.sh',
                      '#!/bin/bash',
                      'echo "$1"',
                      'echo "$2" >&2',
                      mode=0o744)  # -rwxr--r--

    args = ['Hello,', 'World!']

    command = Script(script, *args)
    command.execute()

    result = command.get_result()

    assert result is not None
    assert result.exit_status == 0
    assert result.stdout == 'Hello,\n'
    assert result.stderr == 'World!\n'
