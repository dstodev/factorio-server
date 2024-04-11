# Game server environment

## Requirements

- [Docker](https://docs.docker.com/get-docker/)

All host scripts are tested on `Ubuntu 22.04.3 LTS`

## Configure server

After starting for the first time, server configuration files are created in:

- `./server-files/` (root)
- `./server-files/config` (mod configs)
- `./server-files/server.properties` (server settings)

Edit these files to configure the server.

## Ports

The server uses the following ports:

- `UDP 25565` (Game)
- `TCP 25575` (RCON)

These ports are configurable in `./docker/.env`.

## Start server

You will need a Forge installer for your required version in `./cfg/`.  
You can get one [here](https://files.minecraftforge.net/net/minecraftforge/forge/),
and should look like e.g. `./cfg/forge-1.12.2-14.23.5.2860-installer.jar`

This script will download the server files and start the server. After starting
for the first time, you do not need to use the `--update` flag unless you are
updating the server files.  
`./script/start-server.sh --update`

## Stop server

This script will save the world and gracefully stop the server.  
`./script/stop-server.sh`

This script requires RCON (see below).

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

To use `send-rcon.sh`, you must first set an RCON password.  
**Without an RCON password, you cannot**:

- Gracefully stop the server using `./script/stop-server.sh`,
  because it uses RCON to send the shutdown command.
- Use `./script/backup.sh` without the `--force` flag, because it uses RCON to
  send the save command.

To set an RCON password, edit `./server-files/server.properties`:

- set `enable-rcon=true`
- set `rcon.port=25575` (or update `./docker/.env` to match)
- set `rcon.password=YourRconPassword`

then make a file `./rcon/secret` containing the same password in plaintext.

## Backups

This script creates a backup of important server files to `./backups/`:  
`./script/backup.sh --force`

Backups are created automatically when the server is stopped via `stop-server.sh`.

You should schedule automatic backups using e.g. a cron job. See `./script/backup.sh`
for details.

### Restore from backup

To restore from backup, unzip the backup file you want:  
`tar -xjf ./backups/backup-timestamp.tar.bz2`

Replace the files in `./server-files/world` with the files from the backup.

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
