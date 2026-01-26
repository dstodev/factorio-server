# Config notes

- [Official guide](https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual)

## Init

```shell
./mg.sh hytale stop -- shelf -- download

# run downloader & auth
script/dev-shell.sh hytale
./hytale-downloader-linux-amd64
unzip *.zip
cd Server
java -jar HytaleServer.jar --assets ../Assets.zip
/auth login device
/auth persistence Encrypted
/op add 683ec0a7-9aae-4794-8024-389d2beb6a36
stop
exit

./mg.sh hytale start
```

```shell
script/dev-shell.sh hytale
cd Server
java -jar HytaleServer.jar --help
```

## Updates

```shell
./mg.sh hytale stop -- backup

# run downloader & auth
script/dev-shell.sh hytale
rm *.zip
./hytale-downloader-linux-amd64
unzip *.zip
cd Server
java -jar HytaleServer.jar --assets ../Assets.zip
/auth login device
/auth persistence Encrypted
/op add 683ec0a7-9aae-4794-8024-389d2beb6a36
stop
(set -x; java -jar HytaleServer.jar --help) >help.txt 2>&1
exit

./mg.sh hytale start
```

script/dev-shell.sh hytale

???

./mg.sh hytale start

## Other

ingame commands:

```shell
# Change default world
/world setdefault Nerdhaven
/perm group add Adventure hytale.teleport.command.tp
/perm group add Adventure hytale.command.teleport.world
```
