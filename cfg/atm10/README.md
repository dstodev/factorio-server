# All the Mods 10

## Useful Commands

### Cron Backups

```shell
# Add to crontab, running every day at 5am:
crontab -e

# Update this path to the absolute mg.sh path or relative from $HOME
0 5 * * * repo/mg.sh atm10 rcon --say "Server is restarting!" -- stop -- backup -- start
50 4 * * * repo/mg.sh atm10 rcon --say "Server will restart in 10 minutes!"
59 4 * * * repo/mg.sh atm10 rcon --say "Server will restart in 1 minute!"
```

### Setup QoL rules

```shell
# get
m1/py.sh -- -m cli atm10 rcon --send gamerule playersSleepingPercentage

# set
m1/py.sh -- -m cli atm10 rcon --send gamerule playersSleepingPercentage 50
```

### Reimburse item to player

```shell
m1/py.sh -- -m cli atm10 rcon --send give nynja powah:thermo_generator_basic 2
```

### Increase force loaded chunk limit

- All command nodes listed in:

  `server-files/atm10/hot/world/serverconfig/ftbranks/README.txt`

```shell
# Add player to admin rank
./mg.sh atm10 rcon --send ftbranks add @p admin

# Add extra for specific player
./mg.sh atm10 rcon --send ftbchunks admin extra_force_load_chunks @p add 75

# Reset extra for player
./mg.sh atm10 rcon --send ftbchunks admin extra_force_load_chunks @p set 0

# View ranks
./mg.sh atm10 rcon --send ftbranks list_all_ranks
./mg.sh atm10 rcon --send ftbranks list_ranks_of @p

# List nodes (permissions) of a rank
./mg.sh atm10 rcon --send ftbranks node list admin
./mg.sh atm10 rcon --send ftbranks node list vip
./mg.sh atm10 rcon --send ftbranks node list member

# Manage nodes
# see: server-files/atm10/hot/config/ftbchunks-world.snbt
#      server-files/atm10/hot/world/serverconfig/ftbranks/ranks.snbt
# and: https://docs.feed-the-beast.com/docs/mods/suite/Ranks/Commands
#      https://docs.feed-the-beast.com/docs/mods/suite/Chunks/commands
./mg.sh atm10 rcon --send ftbranks node add member ftbchunks.max_force_loaded 100
./mg.sh atm10 rcon --send ftbranks node remove member ftbchunks.max_force_loaded
./mg.sh atm10 rcon --send ftbranks reload  # might reset changes if not yet written
```

## Server Update Strategy

- [Mod page](https://www.curseforge.com/minecraft/modpacks/all-the-mods-10)

### Steps

```shell
./mg.sh atm10 stop
./mg.sh atm10 backup
./mg.sh atm10 shelf

(update URL in cfg/atm10/download.sh)

./mg.sh atm10 download
./mg.sh atm10 start
./mg.sh atm10 stop

script/dev-shell.sh atm10

rm -r world
cp -r /shelf/1/hot/world .
cp -rv /shelf/1/hot/user*.json .
exit

./mg.sh atm10 start
```

#### Reference

> This is copy/pasted from the eternal2 confdir
>
> - [(view)](../eternal2/UPDATE.md)

How to update the your sever!

1. Download the new server files
2. Unzip the new server files
3. Install the new server files as if you were making a new server (including
   installing the forge server)
4. Start the new server and let it start completely (aka finish generating a new
   world) and then stop the server
5. Copy (don't move, make sure you leave the old files as a backup) these files
   from the old server files to the new server files
   1. "world" folder
   2. usercache.json
   3. usernamecache.json
   4. server.properties
   5. user_jvm_args.txt
   6. ops.json
   7. whitelist.json
   8. and any other files or configs you have changed
6. Run your updated server
