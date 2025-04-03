#!/bin/bash
set -euo pipefail

cwd="$(pwd -P)"
this_dir="$(dirname -- "$(readlink -f -- "$0")")"
relative_path="$(realpath --relative-to="$cwd" "$this_dir")"
relative_basename="$relative_path/$(basename "$0")"

help() {
	cat <<-EOF
		This script sets up and provides access to a Python virtual environment.

		Usage: $(basename "$0") [OPTION]... [ -- COMMAND [ARG]... ]
		  -h, --help      Print this message.
		  -v, --verbose   Print verbose messages.

		  -r, --refresh   Reinitialize the virtual environment.
		  -s, --shell     Start a shell in the virtual environment.

		  -- COMMAND [ ARG... ]   Run a command in the environment.

		Examples:
		  $relative_basename -- -m unittest
		  $relative_basename -- -m manage
		  $relative_basename -s -- python -m unittest
		  $relative_basename -s -- manage
		  $relative_basename -s -- python -m manage.cli
		  $relative_basename -r
	EOF
}

canonical=$(
	getopt --name "$(basename "$0")" \
		--options hvrs \
		--longoptions help,verbose,refresh,shell \
		-- "$@"
) || status=$?

if [ "${status-0}" -ne 0 ]; then
	help
	exit 1
fi

eval set -- "$canonical"

VERBOSE="${VERBOSE:-0}"
refresh=0

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
	-s | --shell)
		shell=true
		;;

	--)
		shift # --
		break
		;;
	esac
	shift # option
done

shell="${shell-false}"

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
	# After venv activate, python and pip are available as commands from the venv
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

if "$shell"; then
	shell_cmd=$(cat /proc/$PPID/cmdline | tr '\0' ' ')
	if [ "$#" -eq 0 ]; then
		exec $shell_cmd
	fi
	exec $shell_cmd -c -- "$(printf '%q ' "$@")"
fi

exec python "$@"
