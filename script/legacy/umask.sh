#!/bin/sh
set -euo pipefail

print_umask() {
	dir=$(mktemp -dt check-umask.XXXXXX)
	echo "umask: $(umask)"
	mkdir "$dir/ dir"  # space in directory name for alignment
	touch "$dir/file"
	stat -c '%n: %A (octal %a)' "$dir"/* | sed 's|/.*/||'
	rm -r "$dir"
}

print_umask
