#!/bin/bash
set -euo pipefail

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
server_dir="$(readlink --canonicalize "$script_dir/..")"

# shellcheck disable=SC2046
export $(xargs <"$script_dir/.env")

rcon_port="$RCON_PORT"

if [ $# -gt 2 ]; then
	print_skip_first_two() {
		shift
		shift
		echo '-------------------------'
		echo '  Extra server options:'
		echo ' ' "$@"
		echo '-------------------------'
	}
	print_skip_first_two "$@" # Known that first two are rcon password
fi

# For file permissions:
# file: -rw-rw-r--
#  dir: drwxrwxr-x
umask 0002

if [ ! -d "$server_dir/factorio/saves" ]; then
	factorio/bin/x64/factorio \
		--create "$server_dir/factorio/saves/world.zip" \
		--map-gen-settings "$script_dir/map-gen-settings.json" \
		--map-settings "$script_dir/map-settings.json"
fi

run() {
	factorio/bin/x64/factorio \
		--start-server-load-latest \
		--server-settings "$script_dir/server-settings.json" \
		--rcon-port "$rcon_port" \
		"$@" ||
		status=$?

	if [ "${status-0}" -ne 0 ]; then
		echo "Server exited with status: $status" >&2
	fi
}

pushd "$server_dir"

# Always run at least once
run "$@"

# Run in a loop until a stop file is present
while [ ! -f "$script_dir/stop" ]; do
	sleep 5
	echo "Restarting server..." >&2
	run "$@"
done

popd

rm --force "$script_dir/stop"
