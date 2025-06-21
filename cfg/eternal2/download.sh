#!/bin/bash
set -euo pipefail

hot_dir="$1"

echo "$hot_dir"
pwd

url='https://github.com/adam9899/MC-Eternal-2/releases/download/v1.0/MCE2-Server-Files-1.0.zip'

wget -O "$hot_dir/eternal2.zip" "$url"

unzip -o "$hot_dir/eternal2.zip" -d "$hot_dir"

cd "$hot_dir/MCE2-Server-Files-1.0" || exit 1

java -Xmx4G -jar './forge-1.20.1-47.4.2-installer.jar' --installServer .
