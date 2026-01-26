#!/bin/bash
set -euo pipefail

# Run multiple commands by separating with `--`:
# mg.sh atm10 rcon --say "Server is restarting!" -- stop -- backup -- start

help() {
	cat <<-EOF
		Usage: $0 [opts]... game [command [command-opts]]... [-- [command [command-opts]]]...

		Tool to manage game servers.
		Defers commands to either m1/py.sh (today) or m2 (future).

		-h, --help      Print this message.
		-v, --verbose   Enable verbose output.

		Examples:
		  $0 -v atm10 start
		  $0 atm10 rcon --say 'Server is restarting!' -- stop -- backup -- start
	EOF
}

if [ "$#" -eq 0 ]; then
	help >&2
	exit 1
fi

# Search for the first argument that does not start with a dash and prepend a
# '--' before it.
amend_cli() {
	local arg_str
	local found_positional=false
	for arg in "$@"; do
		case "$arg" in
		-*) ;;
		*)
			if ! $found_positional; then
				found_positional=true
				arg_str+='-- '
			fi
			;;
		esac
		arg_str+="$arg "
	done
	printf '%s' "$arg_str"
}

status=0
getopt --test >/dev/null 2>&1 || status=$?
if [ "$status" -ne 4 ]; then
	echo 'Error: getopt --test failed. This script requires GNU getopt from util-linux.' >&2
	exit 1
fi

status=0
# shellcheck disable=SC2046
cli=$(getopt --name "$(basename -- "$0")" \
	--options hv \
	--longoptions help,verbose \
	-- $(amend_cli "$@")) || status=$?

if [ "$status" -ne 0 ]; then
	help
	exit 1
fi

eval set -- "$cli"

while :; do
	case "$1" in
	-h | --help)
		help
		exit 0
		;;
	-v | --verbose)
		verbose=true
		;;
	--)
		shift # pop '--'
		break
		;;
	esac
	shift # pop option
done

verbose="${verbose-false}"

if $verbose; then
	printf -- 'Args: |%s|\n' "$cli"
fi

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
	mapfile -d '' -t cmd < <(next_command "$@")
	for _ in "${cmd[@]}"; do
		shift # past parsed args
	done
	if [ "${1-}" = '--' ]; then
		shift # past --
	fi
	flag_verbose=
	$verbose && flag_verbose='-v'
	(
		set -x
		"$manage_path_relative" $flag_verbose -- -m cli "$game" "${cmd[@]}"
	)
done
