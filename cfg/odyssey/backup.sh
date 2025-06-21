#!/bin/bash
set -euo pipefail

hot_dir="$1"
backup_dir="$2"

# Wait for the world directory finish being written to
inotifywait --recursive --monitor --event modify --quiet --timeout 4 "$hot_dir/setup/world" || status=$?

if [ "${status-0}" -ne 0 ]; then
	echo "inotifywait returned: $status" >&2
fi

tar -pczf "$backup_dir/world.tar.gz" -C "$hot_dir/setup" world
