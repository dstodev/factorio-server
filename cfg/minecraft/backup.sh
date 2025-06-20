#!/bin/bash
set -euo pipefail

server_dir="$1"
backup_dir="$2"

# Wait for the world directory finish being written to
inotifywait --recursive --monitor --event modify --quiet --timeout 4 "$server_dir/setup/world" || status=$?

if [ "${status-0}" -ne 0 ]; then
	echo "inotifywait returned: $status" >&2
fi

tar -pczf "$backup_dir/world.tar.gz" -C "$server_dir/setup" world
