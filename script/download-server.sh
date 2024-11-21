#!/bin/bash
set -euo pipefail

SERVER_PKG_URL='https://factorio.com/get-download/stable/headless/linux64'

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
source_dir="$(readlink --canonicalize "$script_dir/..")"

output_dir="${1-$source_dir/server-files}"

copy=(rsync --archive --no-compress)
tmp_dir="$script_dir/tmp-package"
pkg_dest="$tmp_dir/package.tar.xz"

cleanup() {
	# tail to show only the deleted directory path itself, not its contents
	rm --recursive --verbose "$tmp_dir" | tail --lines 1
}

mkdir --parents "$tmp_dir/unpack"

if [ ! -f "$pkg_dest" ]; then # Condition is useful when commenting out `rm` line in cleanup() for debugging
	echo 'Downloading server package...'
	wget --continue --output-document="$pkg_dest" --quiet "$SERVER_PKG_URL"
fi

# Always look in server-files, not $output_dir, for the current hash file.
# start-server.sh passes a temporary directory as $output_dir which will never
# contain the hash file.
cur_hash_file="$source_dir/server-files/package.md5"

if [ -f "$cur_hash_file" ]; then
	cur_hash=$(cat "$cur_hash_file")
fi

new_hash=$(md5sum "$pkg_dest" | cut --delimiter ' ' --fields 1)

echo "   Current hash: ${cur_hash-(none)}"
echo "Downloaded hash: ${new_hash-(none)}"

if [ "${cur_hash-}" = "${new_hash-}" ]; then
	echo 'Server files are up-to-date.'
	cleanup
	exit 2
fi

echo 'Extracting server package...'
tar --directory="$tmp_dir/unpack" --extract --file="$pkg_dest" --no-same-permissions --checkpoint=.1000
echo
chmod --recursive g+w "$tmp_dir/unpack/"
# --delete to remove files not in the package (so you must add any additional files after this line)
# --omit-dir-times because setting for server-files/ requires running as its owner (server-user)
"${copy[@]}" --delete --omit-dir-times "$tmp_dir/unpack/" "$output_dir"

md5sum "$pkg_dest" | cut --delimiter ' ' --fields 1 >"$output_dir/package.md5"
cleanup

cfg_dir="$source_dir/cfg"

# Copy new configuration files to top-level config directory
mkdir --parents "$cfg_dir"
"${copy[@]}" "$output_dir/factorio/data/map-gen-settings.example.json" "$cfg_dir/map-gen-settings.json"
"${copy[@]}" "$output_dir/factorio/data/map-settings.example.json" "$cfg_dir/map-settings.json"
"${copy[@]}" "$output_dir/factorio/data/server-settings.example.json" "$cfg_dir/server-settings.json"
