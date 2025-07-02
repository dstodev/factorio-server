#!/bin/bash
set -euo pipefail

hot_dir="$1"

wget -O "$hot_dir/odyssey-3.zip" "https://www.curseforge.com/api/v1/mods/690733/files/5493802/download"

unzip -o "$hot_dir/odyssey-3.zip" -d "$hot_dir"

chmod u+x "$hot_dir/start-server.sh"
