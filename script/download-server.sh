#!/bin/bash
set -euo pipefail

SERVER_PKG_URL='https://factorio.com/get-download/stable/headless/linux64'

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
source_dir="$(readlink --canonicalize "$script_dir/..")"

output_dir="$source_dir/server-files"

pkg_dest="$output_dir/server.tar.xz"

if [ -e "$pkg_dest" ]; then
	exit 0
fi

mkdir --parents "$output_dir/server"

ln --logical --force "$source_dir/docker/.env" "$output_dir/server/.env"
ln --logical --force "$source_dir/cfg/start.sh" "$output_dir/server/start.sh"

ln --logical --force "$source_dir/cfg/map-gen-settings.json" "$output_dir/server/map-gen-settings.json"
ln --logical --force "$source_dir/cfg/map-settings.json" "$output_dir/server/map-settings.json"
ln --logical --force "$source_dir/cfg/server-settings.json" "$output_dir/server/server-settings.json"

echo 'Downloading server package...'
wget --quiet --continue --output-document="$pkg_dest" "$SERVER_PKG_URL"
tar --extract --file="$pkg_dest" --directory="$output_dir"
