#!/bin/bash
set -euo pipefail
# -e : exit on error
# -u : error on unset variable
# -o pipefail : fail on any error in pipe

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
source_dir="$(readlink --canonicalize "$script_dir/..")"
docker_dir="$source_dir/docker"
server_dir="$source_dir/server-files"
rcon_dir="$source_dir/rcon"

# shellcheck disable=SC2046
export $(xargs <"$docker_dir/.env")

server_name="$SERVER_NAME"

help() {
	cat <<-EOF
		Usage: $(basename "$0") [ -u | -o ]
		  -h, --help    Prints this message.
		  -u, --update  Updates the server files before starting the server.
		  -o, --update-only  Updates the server files and exits.
	EOF
}

canonical=$(getopt --name "$(basename "$0")" \
	--options huo \
	--longoptions help,update,update-only \
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
	-u | --update)
		update_server=true
		;;
	-o | --update-only)
		update_server=true
		update_only=true
		;;
	--)
		shift # --
		break
		;;
	esac
	shift # option
done

update_server=${update_server-false}
update_only=${update_only-false}

if $update_server; then
	if ! sudo --non-interactive true 2>/dev/null; then
		echo sudo password required to update server files.
		echo Required to set server file permissions.
		sudo --validate || exit
	fi
fi

logs_dir="$source_dir/logs"
compose_yml="$docker_dir/compose.yml"
compose=(docker compose --file "$compose_yml")

running_container=$(docker container list --filter name="$server_name-server" --quiet)

if [ -n "$running_container" ]; then
	echo Server is already running!
	exit 2
fi

if $update_server; then
	echo Updating server...
	"${compose[@]}" build base

	mkdir --parents "$server_dir/server"

	"$script_dir/fix-permissions.sh"
	"$script_dir/download-server.sh"
fi

if ! $update_only; then
	if [ ! -f "$rcon_dir/secret" ]; then
		date +%y%m%d%H%M%S%N | md5sum | cut -d ' ' -f 1 >"$rcon_dir/secret"
	fi

	if [ ! -f "$server_dir/server/secret" ]; then
		ln --logical --force "$rcon_dir/secret" "$server_dir/server/secret"
	fi

	compose_run=("${compose[@]}" --progress plain run --rm --service-ports server "$@")

	log="$logs_dir/log-$(date +%Y%j-%H%M%S).log"
	mkdir --parents "$(dirname "$log")"

	echo "Output logging to file: $log"
	echo "Running in screen daemon: screen -r $server_name"

	screen -UdmS "$server_name" -L -Logfile "$log" "${compose_run[@]}"
fi
