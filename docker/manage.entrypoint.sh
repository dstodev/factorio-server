#!/bin/bash
set -euo pipefail

help() {
	cat <<-EOF
		This script sets up the Python virtual environment for the
		manage program, and forwards arguments to it.

		Usage: $(basename "$0") [ -t ] [ -u ] [--]
		  -h, --help    Print this message.
		  -t, --test    Run manage tests, then exit.
		  -u, --update  Update the manage project first.
		  -- [ ... ]    Pass all arguments after -- to the manage program.

		Try:
		  script/manage.sh --test
		  script/manage.sh -- --help
	EOF
}

canonical=$(getopt --name "$(basename "$0")" \
	--options htu \
	--longoptions help,test,update \
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
	-t | --test)
		test_manage=true
		;;
	-u | --update)
		update_manage=true
		;;
	--)
		shift # --
		break
		;;
	esac
	shift # option
done

test_manage=${test_manage-false}
update_manage=${update_manage-false}

pushd manage >/dev/null

venv_activate() {
	# shellcheck source=manage/.venv/bin/activate
	source .venv/bin/activate
}

venv_init() {
	python -m venv .venv &&
		venv_activate &&
		pip install --upgrade pip &&
		pip install --editable .
}

umask 0002

if $update_manage; then
	clean_files=(
		.venv
		build
		manage.egg-info
	)
	for file in "${clean_files[@]}"; do
		if [ -e "$file" ]; then
			rm -r "$file"
		fi
	done
fi

if [ -d .venv ] && ! $update_manage; then
	venv_activate
else
	venv_init # calls venv_activate
fi

if $test_manage; then
	python -m unittest || status=$?
	exit "${status-0}"
fi

manage "$@"

popd >/dev/null
