#!/bin/bash
set -euo pipefail

hot_dir="$1"
backup_dir="$2"

world_dir="$hot_dir/save-data"
world_parent_dir="$(dirname "$world_dir")"

# Wait for the world directory to stop being modified
inotifywait --recursive --monitor \
	--event modify \
	--event create \
	--event delete \
	--event move \
	--event attrib \
	--quiet --timeout 1 "$world_dir" || status=$?

if [ "${status-0}" -ne 0 ]; then
	echo "inotifywait returned: $status" >&2
fi

sync

tar -pczf "$backup_dir/world.tar.gz" -C "$world_parent_dir" "$(basename "$world_dir")"
