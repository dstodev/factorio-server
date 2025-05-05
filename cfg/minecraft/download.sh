#!/bin/bash
set -euo pipefail

server_dir="$1"

echo "$server_dir"
pwd

url='https://www.curseforge.com/api/v1/mods/690733/files/5493802/download'

wget -O "$server_dir/odyssey-3.zip" "$url"

unzip -o "$server_dir/odyssey-3.zip" -d "$server_dir"

chmod u+x "$server_dir/start-server.sh"
