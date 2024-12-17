#!/bin/bash
set -euo pipefail

# Add to crontab, restarting every day at 5am:
#  crontab -e
#  0 5 * * * /absolute/path/to/script/stop-server.sh --restart --time 30 >/dev/null 2>&1

help() {
	cat <<-EOF
		Usage: $(basename "$0") [ -f ] [ -t <time> ] [ -r ] [--]
		  -h, --help     Print this message.
		  -f, --force    Force stop the server. Does not backup or wait.
		  -t <time>, --time <time>
		                 Time in seconds to wait before stopping the server.
		  -r, --restart  Restart the server after stopping.
		  -- [ ... ]     When restarting, pass all arguments after -- to start-server.sh.
	EOF
}

canonical=$(getopt --name "$(basename "$0")" \
	--options hft:r \
	--longoptions help,force,time:,restart \
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
	-f | --force)
		force=true
		;;
	-t | --time)
		if [ -n "$2" ] && [ "$(cut -c 1 <<<"$2")" != '-' ]; then
			time="$2"
			shift # value
		else
			echo 'Error: option --time requires a value.' >&2
			exit 1
		fi
		;;
	-r | --restart)
		restart=true
		;;
	--)
		shift # --
		break
		;;
	esac
	shift # option
done

force=${force-false}
restart=${restart-false}
time=${time-10} # Wait 10 seconds by default

script_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink --canonicalize "$script_dir/..")"

docker_dir="$source_dir/docker"
server_dir="$source_dir/server-files"

rcon="$script_dir/send-rcon.sh"

set -o allexport
# shellcheck source=docker/.env
source "$docker_dir/.env"
set +o allexport

server_name="${SERVER_NAME-game}"

running_container=$(docker container list --filter name="$server_name-server" --quiet)

compose_yml="$docker_dir/compose.yml"
compose=(docker compose --file "$compose_yml")

status=0

if [ -n "$running_container" ]; then
	if $force; then
		"${compose[@]}" down --remove-orphans
	else
		# Give any players time to prepare & leave by themselves
		"$rcon" "Server will stop in $time seconds!" >/dev/null 2>&1 || status=$?

		if [ "$status" -eq 0 ]; then
			# Server responded to rcon command; gracefully terminate
			printf 'Waiting %d seconds before issuing stop command...\n' "$time"
			sleep "$time"
			printf 'Stopping server...\n'

			# Save the server
			"$rcon" '/server-save' >/dev/null 2>&1 || status=$?

			if [ "$status" -ne 0 ]; then
				echo 'Failed to save server with RCON!' >&2
				status=0 # Reset status for next command
			fi

			# Create a stop file to avoid restarting the server (see cfg/start.sh)
			touch "$server_dir/server/stop"

			# Quit the server
			# Even if save failed, /quit will also try to save
			# (but maybe only to autosave file, check logs!)
			"$rcon" '/quit' >/dev/null 2>&1 || status=$?

			if [ "$status" -eq 0 ]; then
				printf 'Waiting for server to close... '
				docker container wait "$running_container" >/dev/null
				printf 'done!\n'
			else
				echo 'Failed to stop server with RCON!' >&2
			fi
		else
			echo 'Server is not responding to RCON!' >&2
		fi

		if [ "$status" -ne 0 ]; then
			echo 'Failed to terminate gracefully!' >&2
			printf 'Stopping container... '
			docker container stop "$running_container" >/dev/null
			printf 'done!\n'
		fi

		"$script_dir/backup.sh" --force # Just stopped server; force to ignore expected rcon failure
	fi
else
	echo 'Server is not running!'

	if ! $restart; then
		exit 2
	fi
fi

if $restart; then
	"$script_dir/start-server.sh" "$@"
fi
