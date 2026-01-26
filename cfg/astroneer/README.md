# Config notes

- [Official guide](https://blog.astroneer.space/p/astroneer-dedicated-server-details/)

## Setup

```shell
./mg.sh astroneer stop -- shelf -- download -- start
script/dev-shell.sh astroneer
vim Astro/Saved/Config/WindowsServer/AstroServerSettings.ini
```

## Common

```shell
./mg.sh astroneer stop -- shelf -- download
./mg.sh astroneer stop -- start
script/vnc.sh astroneer
script/dev-shell.sh astroneer
```

## Updates

```shell
./mg.sh astroneer backup
./mg.sh astroneer shelf
./mg.sh astroneer download
```

script/dev-shell.sh astroneer

???

./mg.sh astroneer start
