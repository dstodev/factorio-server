# Game server environment

## Requirements

- [Docker](https://docs.docker.com/get-docker/)

All host scripts are tested on `Ubuntu 22.04.3 LTS`

## Configure server

Configuration files are located in:

- `./cfg/`

After starting for the first time, server files are created in:

- `./server-files/`

## Ports

The server uses the following ports:

- `UDP 34197` (Game)
- `TCP 34207` (RCON)

These ports are configurable in `./docker/.env`.

## Start server

### For the first time

To start the server for the first time, run:  
`./script/start-server.sh --update-only`

to download the server files without starting the server.

Then, update configuration files in `./cfg` and start the server normally:

### Normally

To start the server normally, run:  
`./script/start-server.sh`

### With updated server files

To update the server before starting, run:  
`./script/start-server.sh --update --restore`

to move the current server files to `./shelf` and update the server files.
After updating, `--restore` will copy world and configuration files from the
latest-"shelved" server files to the new server instance. After, the server
starts normally.

## Stop server

This script will save the world and gracefully stop the server.  
`./script/stop-server.sh`

Gracefully stopping the server requires RCON (see below).

To forcefully stop the server, pass the `--force` flag:  
`./script/stop-server.sh --force`

this will stop the server without saving the world or backing up server files.

## Restart server

To restart the server:  
`./script/stop-server.sh --restart`

You may schedule automatic restarts using e.g. a cron job.
See `./script/stop-server.sh` for details.

## RCON

This environment supports RCON for sending commands to the server:  
`./script/send-rcon.sh MyRconCommand`

This script assumes the server is accessible via `localhost`, but the
underlying C++ client `./rcon/main.cxx` supports sending messages to any host.

## Backups

This script creates a backup of important server files to `./backups/`:  
`./script/backup.sh --force`

Backups are created automatically when the server is stopped via `stop-server.sh`.

You should schedule automatic backups using e.g. a cron job. See `./script/backup.sh`
for details.

### Restore from backup

To restore from backup, unzip the backup file you want:  
`tar -xjf ./backups/backup-timestamp.tar.bz2`

Replace contents of `./server-files/` with files from the backup.

## Other commands

- Rebuild Docker image:  
  `docker compose -f docker/compose.yml build --no-cache server`
- Browse the Docker container with a bash shell:  
  `docker compose -f docker/compose.yml run --entrypoint /bin/bash server`
- Attach to the server screen:  
  `screen -r <server_name>`  
  Keybinds:
  - `CTRL+A` then `D` to detach
  - `CTRL+C` to stop the server
- View disk usage:  
  `du -hs * | sort -hr`

<!-- line break -->

- Attach most-recent log file to the terminal:  
  `./attach-latest-log.sh`

  Keybinds:
  - `SHIFT+F` while attached to await more log text (useful while server is
              still running)
  - `CTRL+C` while awaiting log text to stop awaiting & resume navigation
  - `Q` while navigating to exit logfile

## Permissions

Running `./script/start-server.sh --update` will cause:

- Group `server-group` exists
- User `server-user` (in group `server-group`) exists
- Host user added to group `server-group` *
- Server files owned by `server-user:server-group`
- Repo files owned by `server-group` (for permission to e.g. write backups to `./backups`)

> \* Changes to a user's group membership will not take effect until e.g. the
> user logs out and back in again.
