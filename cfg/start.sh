#!/bin/bash
set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && env pwd --physical)"

# For file permissions:
# file: -rw-rw-r--
#  dir: drwxrwxr-x
umask 0002

run() {
	./run.sh --nogui
}

pushd "$script_dir/.."

# Always run at least once
run

# Run in a loop until a stop file is present
while [ ! -f "$script_dir/stop" ]; do
	run
done

popd

rm --force "$script_dir/stop"
