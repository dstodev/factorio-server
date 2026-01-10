# M2 Project File Structure

## Directories

To organize game server files, the following structure is rooted in the repo
root directory:

```text
server-files/
 ╷
 ├╴game-1/
 │  ╷
 │  ├╴cfg/
 │  │  ├╴config.json
 │  │  ├╴server.dockerfile
 │  │  : ...
 │  │
 │  ├╴script/
 │  │  ├╴download.sh
 │  │  ├╴backup.sh
 │  │  ├╴start.sh
 │  │  : ...
 │  │
 │  ├╴backup/
 │  │  ├╴2025-11-27_20-17-33Z/
 │  │  :  ├╴world.tgz
 │  │     : ...
 │  │
 │  └╴log/
 │     ├╴2025-11-27_20-17-33Z.log
 │     : ...
 │
 ├╴game-2/
 : ├╴cfg/
 ╵ : ...
```
