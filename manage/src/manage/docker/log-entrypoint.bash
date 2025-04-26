#!/bin/bash

# This script runs another program, logging its output to a file.

usage() {
	cat <<-EOF
		Usage: $(basename -- "$0") logfile CMD [ARG...]
	EOF
}

if [ "$#" -lt 2 ]; then
	usage >&2
	exit 1
fi

logfile="$1"
shift # past logfile

# If the first parameter is not a program, ask if the user forgot the logname
if ! command -v "$1" >/dev/null 2>&1; then
	cat <<-EOF >&2
		Error:
		  Logfile name: $logfile
		  Invalid command: $*
	EOF
	usage >&2
	exit 1
fi

exec -- "$@" \
	1> >(sed -u 's/^/stdout: /' | tee -a -- "$logfile" >&1) \
	2> >(sed -u 's/^/stderr: /' | tee -a -- "$logfile" >&2)
