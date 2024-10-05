#!/bin/bash
set -euo pipefail

# Add to crontab, restarting every day at 5am:
#  crontab -e
#  0 5 * * * /absolute/path/to/script/stop-server.sh --restart --time 30 >/dev/null 2>&1

help() {
	cat <<-EOF
		Usage: $(basename "$0") [ -f ] [ -t <time> ] [ -r ]
		  -h, --help     Prints this message.
		  -f, --force    Force stop the server. Does not backup or wait.
		  -t, --time <time>  Time in seconds to wait before stopping the server.
		  -r, --restart  Restart the server after stopping.
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
		if [ -n "$2" ] && [ "${2:0:1}" != "-" ]; then
			time="$2"
			shift # value
		else
			echo "Error: option --time requires a value." >&2
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

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
source_dir="$(readlink --canonicalize "$script_dir/..")"
docker_dir="$source_dir/docker"
server_dir="$source_dir/server-files"

compose_yml="$docker_dir/compose.yml"
compose=(docker compose --file "$compose_yml")

rcon="$script_dir/send-rcon.sh"

# shellcheck disable=SC2046
export $(xargs <"$docker_dir/.env")

server_name="$SERVER_NAME"

running_container=$(docker container list --filter name="$server_name-server" --quiet)

if [ -n "$running_container" ]; then
	if $force; then
		"${compose[@]}" down --remove-orphans
	else
		# Create a stop file to signal the server to stop (see cfg/start.sh)
		touch "$server_dir/server/stop"

		# Give any players time to gracefully leave
		printf 'Waiting %d seconds before issuing stop command...\n' "$time"
		"$rcon" "Server will stop in $time seconds!"
		sleep "$time"
		printf 'done!\n'

		"$rcon" "/quit"
		printf 'Waiting for server to close... '
		docker container wait "$running_container" >/dev/null
		printf 'done!\n'

		"$script_dir/backup.sh" --force # Just stopped server; force to ignore expected rcon failure
	fi
else
	echo Server is not running!

	if ! $restart; then
		exit 2
	fi
fi

if $restart; then
	"$script_dir/start-server.sh"
fi
