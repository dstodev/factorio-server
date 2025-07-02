# Manage2

The second version of `manage`, after the Python version.

This version is written in Go as its Docker API seems more fleshed out.
(Makes sense since Docker is written in Go)

This version is structured differently:

Rather than composing behaviors around a stateful GameContainer class, manage2
uses more direct functional calls to the Docker API to perform actions on
containers

Here are some example invocations, assuming you're running for Factorio:

```shell
m2 --help
m2 factorio status
m2 factorio download
m2 factorio start
m2 factorio stop
m2 factorio restart
m2 factorio backup
m2 factorio update
m2 factorio shelf
m2 factorio stdin commands-to-send
m2 factorio rcon commands-to-send
m2 factorio say message-to-say

```

## Guidelines Followed

- [Project structure](https://appliedgo.com/blog/go-project-layout)
