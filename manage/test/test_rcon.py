import unittest

from manage.rcon import Client


class TestRcon(unittest.TestCase):
    def test_rcon(self):
        client = Client()
        self.assertIsNotNone(client)
