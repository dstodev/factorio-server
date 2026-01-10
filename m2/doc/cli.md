# Manage2 CLI

The Manage2 (m2) Command-Line Interface (CLI) provides a simple, common toolset
for managing many unrelated game servers.

## Common Usage

```text
m2 <game> <verb> [verb-options]
```

e.g.:

```shell
m2 minecraft start
m2 minecraft backup
m2 minecraft rcon 'give @a diamond_sword'
```

## Observers

`m2 <game> observe [observe-options]`

### Options

- `--script <path>`

  Copy the script content at `path` into a temporary file within the server
  container, then run it.

  Server log output is replicated to all observers registered at the time the
  output originates. New observers do not receive prior output.

---
