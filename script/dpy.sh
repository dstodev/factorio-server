#!/bin/bash
set -euo pipefail

# Running Python in a Docker container has a few advantages:
# - Easy to control runtime environment and dependencies regardless of host system configuration
# - Useful to test OS-facing features in a controlled environment without affecting the host system
# - Allows us to leverage Docker's service networking features
# - Helps keep host system clean by storing __pycache__ and other artifacts in the container
#
# Inside the container, the user's home directory is persisted by a Docker volume.
# Use this script's --refresh option to clear it. (see docker/compose.yml)
#
# Example usage:
#   script/dpy.sh -- python -m unittest discover -s manage
#   script/dpy.sh -- manage --help
#
# Related & supporting files:
# - docker/compose.yml (py service)
# - docker/py.dockerfile
# - docker/py.entrypoint.sh

help() {
	cat <<-EOF
		This script sets up and provides access to a Python virtual environment
		inside a Docker container.

		Usage: $(basename "$0") [ -r[r] ] [ -- COMMAND [ ARG... ] ]
		  -h, --help      Print this message.
		  -r, --refresh   Clear persistent data & apply Dockerfile changes.
		                  Provide twice to completely rebuild the image.

		  -- COMMAND [ ARG... ]   Run a command in the environment.
	EOF
}

canonical=$(getopt --name "$(basename "$0")" \
	--options hru \
	--longoptions help,refresh,update \
	-- "$@") || status=$?

if [ "${status-0}" -ne 0 ]; then
	help
	exit 1
fi

eval set -- "$canonical"

refresh=0

while :; do
	case "$1" in
	-h | --help)
		help
		exit 0
		;;
	-r | --refresh)
		refresh=$((refresh + 1))
		;;
	--)
		shift # --
		break
		;;
	esac
	shift # option
done

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink -f -- "$this_dir/..")"

docker_dir="$source_dir/docker"
compose_yml="$docker_dir/compose.yml"

compose=(docker compose --file "$compose_yml")

if [ "$refresh" -gt 0 ]; then
	"${compose[@]}" down --remove-orphans --volumes py

	build_cmd=("${compose[@]}" build)
	[ "$refresh" -gt 1 ] && build_cmd+=(--no-cache)
	"${build_cmd[@]}" py
fi

"${compose[@]}" run --rm py "$@"
