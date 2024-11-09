#!/bin/bash
set -euo pipefail
# -e : exit on error
# -u : error on unset variable
# -o pipefail : fail on any error in pipe
# Docs: https://www.gnu.org/software/bash/manual/html_node/The-Set-Builtin.html

help() {
	cat <<-EOF
		Usage: $(basename "$0") [ -u | -o ]
		  -h, --help    Print this message.
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

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
source_dir="$(readlink --canonicalize "$script_dir/..")"

docker_dir="$source_dir/docker"

# shellcheck disable=SC2046
export $(xargs <"$docker_dir/.env")

server_name="${SERVER_NAME-game}"

running_container=$(docker container list --filter name="$server_name-server" --quiet)

if [ -n "$running_container" ]; then
	echo Server is already running!
	exit 2
fi

compose_yml="$docker_dir/compose.yml"
compose=(docker compose --file "$compose_yml")
server_dir="$source_dir/server-files"

if $update_server; then
	echo Updating server...

	"${compose[@]}" build base

	umask 0002
	# For permissions:
	# file: -rw-rw-r-- (0666 & ~0002 = 0664)
	#  dir: drwxrwxr-x (0777 & ~0002 = 0775)
	#
	# to test umask:
	# echo "umask: $(umask)" && rm -rf /tmp/check-umask && mkdir -p '/tmp/check-umask/ dir' && touch /tmp/check-umask/file && stat -c '%n: %A (octal %a)' /tmp/check-umask/* | sed 's|/.*/||g' && rm -r /tmp/check-umask

	mkdir --parents "$source_dir/cfg"
	mkdir --parents "$source_dir/backups"

	"$script_dir/shelve-state.sh"
	"$script_dir/init-permissions.sh" # Set up host environment permissions
	"$script_dir/download-server.sh"  # Download files with group set from setgid
fi

if $update_only; then
	exit 0
fi

if [ ! -d "$server_dir" ]; then
	echo 'Server files not found. Run with --update to acquire them.' >&2
	exit 3
fi

mkdir --parents "$server_dir/server"

link=(ln --logical --force)

"${link[@]}" "$source_dir/docker/.env" "$server_dir/server/.env"
"${link[@]}" "$source_dir/cfg/start.sh" "$server_dir/server/start.sh"

"${link[@]}" "$source_dir/cfg/map-gen-settings.json" "$server_dir/server/map-gen-settings.json"
"${link[@]}" "$source_dir/cfg/map-settings.json" "$server_dir/server/map-settings.json"
"${link[@]}" "$source_dir/cfg/server-settings.json" "$server_dir/server/server-settings.json"

rcon_dir="$source_dir/rcon"

if [ ! -f "$rcon_dir/secret" ]; then
	date +%y%m%d%H%M%S%N | md5sum | cut -d ' ' -f 1 >"$rcon_dir/secret"
fi

rcon_password=$(<"$rcon_dir/secret")

if [ -f "$server_dir/server/stop" ]; then
	rm "$server_dir/server/stop"
	echo 'Removed previous stopfile.'
fi

logs_dir="$source_dir/logs"

mkdir --parents "$logs_dir"

log="$logs_dir/log-$(date +%Y%j-%H%M%S).log"

echo "Output logging to file: $log"
echo "Running in screen daemon: screen -r $server_name"

compose_run=(
	"${compose[@]}"
	--progress plain
	run
	--rm
	--service-ports
	server
	--rcon-password "$rcon_password"
	"$@"
)

screen -UdmS "$server_name" -L -Logfile "$log" "${compose_run[@]}"
