#!/bin/bash
set -euo pipefail

# Top-most entrypoint to manage game servers.
# Defers to either m1/py.sh (today) or m2 (future).
#
# Run multiple commands by separating with `--`:
# mg.sh atm10 rcon --say "Server is restarting!" -- stop -- backup -- start

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
manage_path="$this_dir"/m1/py.sh

game="$1"
shift

# Use relative path for nicer printout below where set -x is enabled.
manage_path_relative="$(realpath --relative-to="$(pwd -P)" "$manage_path")"

have_more_args() {
	[ "$#" -gt 0 ]
}

next_command() {
	local args=()
	while have_more_args "$@"; do
		arg="$1"
		shift
		if [ "$arg" = '--' ]; then
			break
		fi
		args+=("$arg")
	done
	printf '%s\0' "${args[@]}"
}

while have_more_args "$@"; do
	mapfile -d '' -t args < <(next_command "$@")
	for _ in "${args[@]}"; do
		shift # past parsed args
	done
	if [ "${1-}" = '--' ]; then
		shift # past --
	fi
	(
		set -x
		"$manage_path_relative" -- -m cli "$game" "${args[@]}"
	)
done
