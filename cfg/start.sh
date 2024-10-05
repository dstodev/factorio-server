#!/bin/bash
set -euo pipefail

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
server_dir="$(readlink --canonicalize "$script_dir/..")"

# shellcheck disable=SC2046
export $(xargs <"$script_dir/.env")

rcon_port="$RCON_PORT"
rcon_password="$(cat "$script_dir/secret")"

# For file permissions:
# file: -rw-rw-r--
#  dir: drwxrwxr-x
umask 0002

if [ $# -gt 0 ]; then
	echo "Options:" "$@"
fi

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
		--rcon-password "$rcon_password" \
		"$@"
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
