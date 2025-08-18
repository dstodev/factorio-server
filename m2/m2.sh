#!/bin/bash
set -euo pipefail

reset

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
build_dir="$this_dir/bin"
bin='m2'

out="$build_dir/$bin"
go build -o "$out" "$this_dir"

cd "$build_dir"
echo "$bin" "$@"
"$out" "$@"
