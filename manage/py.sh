#!/bin/bash
set -euo pipefail

cwd="$(pwd -P)"
this_dir="$(dirname -- "$(readlink -f -- "$0")")"
relative_path="$(realpath --relative-to="$cwd" "$this_dir")"
relative_basename="$relative_path/$(basename "$0")"

PYTHON_VERSION="${PYTHON_VERSION:-3.10}"

VERBOSE="${VERBOSE:-0}"
export VERBOSE

help() {
	cat <<-EOF
		This script sets up and provides access to a Python virtual environment.

		Usage: $(basename "$0") [OPTION]... [ -- [COMMAND] [ARG]...]
		  -h, --help      Print this message.
		  -v, --verbose   Print verbose messages.

		  -r, --refresh   Reinitialize the virtual environment.
		                  Repeat to fully rebuild the environment.
		  -s, --shell     Start a shell in the virtual environment.

		  -- [COMMAND] [ARG]...   Run a command in the environment.

		Examples:
		  $relative_basename -- -m cli
		  $relative_basename -s -- pytest
		  $relative_basename -vrrs --
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

refresh=0

while :; do
	case "$1" in
	-h | --help)
		help
		exit 0
		;;
	-v | --verbose)
		VERBOSE=$((VERBOSE + 1))
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

venv_stem='.venv'
venv_dir="$this_dir/$venv_stem"
build_dir="$venv_dir/.build"

set -o allexport
PYTHONPYCACHEPREFIX="$venv_dir/.pycache"
PYTEST_ADDOPTS="-o cache_dir=$venv_dir/.pytest_cache"
set +o allexport

cd "$this_dir" || exit 1

# shellcheck source=script/util.sh
source "$source_dir/script/util.sh"

fv="$(flag_verbose)"
fq="$(flag_quiet)"

pretty_rm() {
	rm ${fv:+"$fv"} --force --recursive "$1" | tail --lines 1 | sed 's/^/-- /'
}

clean() {
	if [ $# -gt 0 ]; then
		find "$this_dir" "$@" | while read -r file; do
			pretty_rm "$file"
		done
	fi
}

venv_activate() {
	# shellcheck disable=SC1091
	source "$venv_dir/bin/activate"
}

venv_init() {
	echo "-- Initializing Python virtual environment: $venv_dir"

	# Remove venv_dir from PATH before searching for Python
	# Important if running with -rr from an already-acitvated environment
	PATH="$(echo "$PATH" | tr ':' '\n' | grep -v "$venv_dir" | tr '\n' ':' | sed 's/:$//')"

	py="$(py_interpreter "$PYTHON_VERSION")"
	verbose "-- System interpreter: $py"

	"$py" -m venv "$venv_dir"
	venv_activate

	verbose "-- Activated interpreter: $(which python)"

	# After venv activate, python and pip are available as commands from the venv
	pip ${fq:+"$fq"} install --upgrade pip

	pretty_rm "$build_dir"

	find "$source_dir" -type f -name pyproject.toml | while read -r pyproject; do
		verbose "-- Installing project: $pyproject"
		dir="$(dirname -- "$pyproject")"
		pip ${fq:+"$fq"} install --editable "$dir"'[dev]'
	done
}

if [ "$refresh" -gt 1 ]; then
	clean -type d -name '*.egg-info'
	pretty_rm "$venv_dir"
fi

if [ -d "$venv_dir/bin" ] && [ "$refresh" -eq 0 ]; then
	venv_activate
else
	venv_init # calls venv_activate
fi

if "$shell"; then
	shell_cmd='/bin/bash'
	rcfile_ps1="export PS1='($venv_stem) \$(basename \"\$(pwd)\")\$ '"

	if [ "$#" -eq 0 ]; then
		exec $shell_cmd --rcfile <(echo "$rcfile_ps1") -i
	fi
	exec $shell_cmd -c -- "$(printf '%q ' "$@")"
fi

exec python "$@"
