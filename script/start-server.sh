#!/bin/bash
set -euo pipefail
# -e : exit on error
# -u : error on unset variable
# -o pipefail : fail on any error in pipe

script_dir="$(cd "$(dirname "$0")" && env pwd --physical)"
source_dir="$(readlink --canonicalize "$script_dir/..")"
docker_dir="$source_dir/docker"
server_dir="$source_dir/server-files"

# shellcheck disable=SC2046
export $(xargs <"$docker_dir/.env")
server_name="$SERVER_NAME"

help() {
	cat <<-EOF
		Usage: $(basename "$0") [ -u ]
		  -h, --help    Prints this message.
		  -u, --update  Updates the server files before starting the server.
		  -o, --update-only  Updates the server files and exits.
	EOF
}

canonicalized=$(getopt --name "$(basename "$0")" \
	--options huo \
	--longoptions help,update,update-only \
	-- "$@") || status=$?

if [ "${status-0}" -ne 0 ]; then
	help
	exit 1
fi

eval set -- "$canonicalized"

for arg in "$@"; do
	case $arg in
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
	esac
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

	"$script_dir/fix-permissions.sh"

	mkdir --parents "$server_dir/server"
	ln --logical --force "$source_dir/cfg/start.sh" "$server_dir/server/start.sh"

	cp "$source_dir/cfg/forge-"*"-installer.jar" "$server_dir/installer.jar"
	"${compose[@]}" run cmd "java -jar installer.jar --installServer"
fi

if ! $update_only; then
	compose_run=("${compose[@]}" --progress plain run --rm --service-ports server)

	log="$logs_dir/log-$(date +%Y%j-%H%M%S).log"
	mkdir --parents "$(dirname "$log")"

	echo Running command: "${compose_run[*]}"
	echo "... with output logging to file: $log"
	echo "... in screen daemon; screen -r $server_name"

	screen -dmS "$server_name" -L -Logfile "$log" "${compose_run[@]}"
fi
