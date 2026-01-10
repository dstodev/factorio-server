#!/bin/bash
set -euo pipefail

hot_dir="$1"
backup_dir="$2"

world_dir="$(find "$hot_dir" -maxdepth 2 -type d -name 'world')"
world_parent_dir="$(dirname -- "$world_dir")"

# Wait for the world directory to stop being modified
inotifywait --recursive --monitor \
	--event modify \
	--event create \
	--event delete \
	--event move \
	--event attrib \
	--quiet --timeout 5 "$world_parent_dir/world" || status=$?

if [ "${status-0}" -ne 0 ]; then
	echo "inotifywait returned: $status" >&2
fi

sync

tar -pczf "$backup_dir/world.tar.gz" -C "$world_parent_dir" world
