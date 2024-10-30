#!/bin/bash
set -euo pipefail

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
source_dir="$(readlink --canonicalize "$script_dir/..")"

name="$source_dir/server-files"

check() {
	[ -e "$1" ] || [ -L "$1" ]
}

if check "$name"; then
	i=1
	while check "$name-$i"; do
		i=$((i + 1))
	done
	new_name="$name-$i"
	echo "Shelving server files: $name => $new_name"
	mv "$name" "$new_name"
fi
