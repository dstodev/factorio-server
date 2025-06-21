#!/bin/bash
set -euo pipefail

hot_dir="$1"
backup_dir="$2"

world_parent_dir="$hot_dir/MCE2-Server-Files-1.0"

# Wait for the world directory finish being written to
inotifywait --recursive --monitor --event modify --quiet --timeout 4 "$world_parent_dir/world" || status=$?

if [ "${status-0}" -ne 0 ]; then
	echo "inotifywait returned: $status" >&2
fi

tar -pczf "$backup_dir/world.tar.gz" -C "$world_parent_dir" world
