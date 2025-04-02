#!/bin/bash
set -euo pipefail

help() {
	cat <<-EOF
		This script sets up and provides access to a Python virtual environment.

		Usage: $(basename "$0") [ -r ] [ -- COMMAND [ ARG... ] ]
		  -h, --help      Print this message.
		  -v, --verbose   Print verbose messages.

		  -r, --refresh   Reinitialize the virtual environment.

		  -- COMMAND [ ARG... ]   Run a command in the environment.
	EOF
}

canonical=$(getopt --name "$(basename "$0")" \
	--options hvru \
	--longoptions help,verbose,refresh,update \
	-- "$@") || status=$?

if [ "${status-0}" -ne 0 ]; then
	help
	exit 1
fi

eval set -- "$canonical"

refresh=0
VERBOSE="${VERBOSE:-0}"

while :; do
	case "$1" in
	-h | --help)
		help
		exit 0
		;;
	-v | --verbose)
		VERBOSE=$((VERBOSE + 1))
		export VERBOSE
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

umask 0002

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink -f -- "$this_dir/..")"

venv_dir="$this_dir/.venv"
build_dir="$venv_dir/.build"

PYTHONPYCACHEPREFIX="$venv_dir/.pycache"
export PYTHONPYCACHEPREFIX

cd "$this_dir" || exit 1

# shellcheck source=script/util.sh
source "$source_dir/script/util.sh"

py="$(py_interpreter 3.10)"
verbose "-- Using python interpreter: $py"

venv_activate() {
	# shellcheck disable=SC1091
	source "$venv_dir/bin/activate"
}

venv_init() {
	echo "-- Initializing python virtual environment: $venv_dir"

	"$py" -m venv "$venv_dir"
	venv_activate
	pip install --upgrade pip

	rm --force --recursive --verbose "$build_dir" | tail --lines 1

	find "$source_dir" -name pyproject.toml | while read -r pyproject; do
		echo "-- Installing project: $pyproject"
		dir="$(dirname -- "$pyproject")"
		stem=$(basename -- "$dir")

		mkdir --parents "$build_dir/$stem"
		ln --relative --symbolic "$dir"/* "$build_dir/$stem"/
		pip install --editable "$build_dir/$stem"
	done
}

if [ -d "$venv_dir/bin" ] && [ "$refresh" -eq 0 ]; then
	venv_activate
else
	venv_init # calls venv_activate
fi

unit() {
	# Helper for running unit tests
	# Run like:
	#   manage/py.sh -- unit .
	# or
	#   manage/py.sh
	#   > (in shell) unit .
	python -m unittest discover --start-directory "$1"
}
export -f unit

if [ "$#" -eq 0 ]; then
	exec /bin/bash
fi

exec /bin/bash -c -- "$(printf '%q ' "$@")"
