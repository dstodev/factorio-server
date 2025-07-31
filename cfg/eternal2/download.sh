#!/bin/bash
set -euo pipefail

hot_dir="$1"

# https://www.curseforge.com/minecraft/modpacks/mc-eternal-2/files/all
wget -O "$hot_dir/eternal2.zip" "https://www.curseforge.com/api/v1/mods/1243287/files/6819570/download"

unzip -o "$hot_dir/eternal2.zip" -d "$hot_dir"

cd "$hot_dir" || exit 1

java -Xmx4G -jar './forge-'*'-installer.jar' --installServer .
