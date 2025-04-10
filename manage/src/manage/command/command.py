class Command:
    def __init__(self):
        self._actions = []

    def add_action(self, action):
        assert hasattr(action, 'execute'), 'Action must have an execute method.'
        self._actions.append(action)

    def execute(self):
        for action in self._actions:
            action.execute()
