# TerraFirmaGreg Modern

## Useful Commands

### Cron Backups

```shell
# Add to crontab, running every day at 5am:
crontab -e

# Update this path to the absolute mg.sh path or relative from $HOME
0 5 * * * repo/mg.sh tfg rcon --say "Server is restarting!" -- stop -- backup -- start
50 4 * * * repo/mg.sh tfg rcon --say "Server will restart in 10 minutes!"
59 4 * * * repo/mg.sh tfg rcon --say "Server will restart in 1 minute!"
```

### Setup QoL rules

```shell
# get
m1/py.sh -- -m cli tfg rcon --send gamerule playersSleepingPercentage

# set
m1/py.sh -- -m cli tfg rcon --send gamerule playersSleepingPercentage 50
```

## Server Update Strategy

https://www.curseforge.com/minecraft/modpacks/terrafirmagreg-modern

### Steps

```shell
m1/py.sh -- -m cli tfg stop
m1/py.sh -- -m cli tfg backup
m1/py.sh -- -m cli tfg shelf

(update URL in cfg/tfg/download.sh)

m1/py.sh -- -m cli tfg download
m1/py.sh -- -m cli tfg start
m1/py.sh -- -m cli tfg stop

script/dev-shell.sh tfg
rm -r world
cp -r /shelf/1/hot/world .
cp -rv /shelf/1/hot/user*.json .
exit

m1/py.sh -- -m cli tfg start
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
