from manage.command.rcon import Client


def test_rcon():
    client = Client()
    assert client is not None
