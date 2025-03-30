#!/bin/bash
set -euo pipefail

umask 0002

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink -f -- "$this_dir/..")"

venv_dir="$HOME/.venv"
build_dir="$HOME/.build"

PYTHONPYCACHEPREFIX="$HOME/.pycache"
export PYTHONPYCACHEPREFIX

venv_activate() {
	# shellcheck disable=SC1091
	source "$venv_dir/bin/activate"
}

venv_init() {
	echo "-- Initializing python virtual environment: $venv_dir"

	python -m venv "$venv_dir" &&
		venv_activate &&
		pip install --upgrade pip
}

if [ -d "$venv_dir/bin" ]; then
	venv_activate
else
	venv_init # calls venv_activate

	find "$source_dir" -name pyproject.toml | while read -r pyproject; do
		echo "-- Installing project: $pyproject"
		dir="$(dirname -- "$pyproject")"
		stem=$(basename -- "$dir")

		mkdir --parents "$build_dir/$stem"
		ln --relative --symbolic "$dir"/* "$build_dir/$stem"/
		pip install --editable "$build_dir/$stem"
	done
fi

if [ -z "$*" ]; then
	exec /bin/bash
fi

exec "$@"
