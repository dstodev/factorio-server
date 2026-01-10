#!/bin/bash
set -euo pipefail

hot_dir="$1"

# https://www.curseforge.com/minecraft/modpacks/all-the-mods-10/files/all
wget -O "$hot_dir/pack.zip" "https://www.curseforge.com/api/v1/mods/925200/files/7285673/download"

unzip -o "$hot_dir/pack.zip" -d "$hot_dir"
rm "$hot_dir/pack.zip"

chmod u+x "$hot_dir/startserver.sh"
