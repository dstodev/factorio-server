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

./mg.sh starrupt download
./mg.sh starrupt backup
./mg.sh starrupt shelf
./mg.sh starrupt download

script/dev-shell.sh starrupt

???

./mg.sh starrupt start
