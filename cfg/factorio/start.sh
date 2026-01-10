#!/bin/bash
set -euo pipefail

# Requires provided inputs:
#   $1 - rcon port
#   $2 - rcon password

# TODO: These are no-longer the correct args
rcon_port="$1"
rcon_password="$2"

shift
shift

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
server_dir="$(readlink -f -- "$this_dir/..")"

if [ $# -gt 0 ]; then
	echo '-------------------------'
	echo '  Extra server options:'
	echo ' ' "$@"
	echo '-------------------------'
fi

umask 0002
# For file permissions:
# file: -rw-rw-r-- (0666 & ~0002 = 0664)
#  dir: drwxrwxr-x (0777 & ~0002 = 0775)

if [ ! -d "$server_dir/factorio/saves" ]; then
	factorio/bin/x64/factorio \
		--create "$server_dir/factorio/saves/world.zip" \
		--map-gen-settings "$this_dir/map-gen-settings.json" \
		--map-settings "$this_dir/map-settings.json"
fi

run() {
	factorio/bin/x64/factorio \
		--start-server-load-latest \
		--server-settings "$this_dir/server-settings.json" \
		--rcon-port "$rcon_port" \
		--rcon-password "$rcon_password" \
		"$@" ||
		status=$?

	if [ "${status-0}" -ne 0 ]; then
		echo "Server exited with status: $status" >&2
	fi
}

pushd "$server_dir"

# Always run at least once
run "$@"

# Run in a loop until a file named "stop" (a "stop file") is present
while [ ! -f "$this_dir/stop" ]; do
	sleep 5
	echo 'Restarting server...' >&2
	run "$@"
done

popd

rm --force --verbose "$this_dir/stop"
