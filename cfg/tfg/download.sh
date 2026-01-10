#!/bin/bash
set -euo pipefail

hot_dir="$1"

# https://www.curseforge.com/minecraft/modpacks/terrafirmagreg-modern/files/all
# https://github.com/TerraFirmaGreg-Team/Modpack-Modern/releases
pkg_url='https://github.com/TerraFirmaGreg-Team/Modpack-Modern/releases/download/0.11.8/TerraFirmaGreg-Modern-0.11.8-serverpack.zip'

wget -O "$hot_dir/tfg.zip" "$pkg_url"
unzip -o "$hot_dir/tfg.zip" -d "$hot_dir"
rm "$hot_dir/tfg.zip"
