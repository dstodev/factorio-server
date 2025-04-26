#!/bin/sh

# This script runs another program, logging its output to a file.

# if [ "$#" -gt 0 ]; then
# 	printf 'Received: %s\n' "$@"
# fi

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

stdout_pipe=$(mktemp -u)
stderr_pipe=$(mktemp -u)
mkfifo "$stdout_pipe" "$stderr_pipe"

nohup sed 's/^/stdout: /' <"$stdout_pipe" | tee -a -- "$logfile" >&1 &
nohup sed 's/^/stderr: /' <"$stderr_pipe" | tee -a -- "$logfile" >&2 &

exec "$@" 1>"$stdout_pipe" 2>"$stderr_pipe"

# This script is intended to run inside a container, so no need to worry about
# cleaning up the named pipes.
