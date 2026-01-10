# Config notes

- (Oct. 9, 2025)
  RCON password is 108 characters long because the server does not support
  rcon over a socket; commands are apparently only issuable via ingame chat.
  To login as admin, send a chat message:

    /AdminLogin rcon-password

  The maximum message size is 120 characters, limiting the RCON password
  length to 108 characters:

  ```text
  /AdminLogin rQOX32kiZioOkEciP9rMU7IY8LquEzHwvvEN3ocmmMi3LPtmEmzfFoNHrOemBxbxhl848Zd0rx4WXSxObwcsrooZnKNDCYhd3uw4QWlwWAzH
  ╷           └╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴ 108 characters ╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶┤
  └╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴╴ 120 characters ╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶╶┘
  ```

## Updates

m1/py.sh -- -m cli icarus stop
m1/py.sh -- -m cli icarus backup
m1/py.sh -- -m cli icarus shelf
m1/py.sh -- -m cli icarus download

script/dev-shell.sh icarus
cp -r /shelf/2/hot/Icarus/Saved /hot/Icarus/

m1/py.sh -- -m cli icarus start
