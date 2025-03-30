#!/bin/bash
set -euo pipefail

# script/dpy.sh -- python -m unittest discover -s manage

help() {
	cat <<-EOF
		This script sets up and provides access to a Python virtual environment
		inside a Docker container.

		Usage: $(basename "$0") [ -- COMMAND [ ARG... ] ]
		  -h, --help    Print this message.
		  -c, --clean   Rebuild the container.

		  -- COMMAND [ ARG... ]   Run a command in the environment.
	EOF
}

canonical=$(getopt --name "$(basename "$0")" \
	--options hcu \
	--longoptions help,clean,update \
	-- "$@") || status=$?

if [ "${status-0}" -ne 0 ]; then
	help
	exit 1
fi

eval set -- "$canonical"

while :; do
	case "$1" in
	-h | --help)
		help
		exit 0
		;;
	-c | --clean)
		clean=true
		;;
	--)
		shift # --
		break
		;;
	esac
	shift # option
done

clean="${clean-false}"

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink -f -- "$this_dir/..")"

docker_dir="$source_dir/docker"
compose_yml="$docker_dir/compose.yml"

compose=(docker compose --file "$compose_yml")

if $clean; then
	"${compose[@]}" down --remove-orphans --volumes py
	"${compose[@]}" build py
fi

"${compose[@]}" run --rm py "$@"
