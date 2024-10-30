#!/bin/bash
set -euo pipefail

SERVER_PKG_URL='https://factorio.com/get-download/stable/headless/linux64'

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
source_dir="$(readlink --canonicalize "$script_dir/..")"
output_dir="$source_dir/server-files"

copy=(rsync --archive --no-compress)
tmp_dir="$script_dir/tmp-package"
cfg_dir="$source_dir/cfg"
pkg_dest="$tmp_dir/package.tar.xz"

mkdir --parents "$tmp_dir/unpack"

if [ ! -f "$pkg_dest" ]; then # Condition is useful when `rm` line is commented out for debugging
	echo 'Downloading server package...'
	wget --continue --output-document="$pkg_dest" --quiet "$SERVER_PKG_URL"
fi

echo 'Extracting server package...'
tar --directory="$tmp_dir/unpack" --extract --file="$pkg_dest" --no-same-permissions
chmod --recursive g+w "$tmp_dir/unpack/"
# --omit-dir-times because setting for server-files/ requires running as its owner (server-user)
"${copy[@]}" --delete --omit-dir-times "$tmp_dir/unpack/" "$output_dir"
rm --recursive "$tmp_dir"

mkdir --parents "$cfg_dir"
"${copy[@]}" "$output_dir/factorio/data/map-gen-settings.example.json" "$cfg_dir/map-gen-settings.json"
"${copy[@]}" "$output_dir/factorio/data/map-settings.example.json" "$cfg_dir/map-settings.json"
"${copy[@]}" "$output_dir/factorio/data/server-settings.example.json" "$cfg_dir/server-settings.json"
