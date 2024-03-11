#!/bin/bash
set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && env pwd --physical)"

# For file permissions:
# file: -rw-rw-r--
#  dir: drwxrwxr-x
umask 0002

# Run in a loop unless a stop file is present
while [ ! -f "$script_dir/stop" ]; do
	java -Xmx8G -Xms4G -jar "$script_dir/../forge-"*".jar" --nogui
done

rm --force "$script_dir/stop"
