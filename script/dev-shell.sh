#!/bin/bash
set -euo pipefail
# -e : exit on error
# -u : error on unset variable
# -o pipefail : fail on any error in pipe
# Docs: https://www.gnu.org/software/bash/manual/html_node/The-Set-Builtin.html

help() {
	cat <<-EOF
	Usage: $0 game [-r]

	Launch a shell in the server environment for a game.

	-h, --help   Print this message.
	-r, --root   Run as root instead of as server's user.
	EOF
}

getopt --test >/dev/null 2>&1 || status=$?
if [ "${status-0}" -ne 4 ]; then
	echo 'Error: getopt --test failed. This script requires GNU getopt from util-linux.' >&2
	exit 1
fi
unset status

args=$(getopt --name "$(basename -- "$0")" \
	--options hr \
	--longoptions help,root \
	-- "$@") || status=$?

if [ "${status-0}" -ne 0 ]; then
	help
	exit 1
fi

eval set -- "$args"

while :; do
	case "$1" in
	-h | --help)
		help
		exit 0
		;;
	-r | --root)
		root=true
		;;
	--)
		shift # --
		break
		;;
	esac
	shift # option
done

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink -f -- "$this_dir/..")"
rcon_dir="$source_dir/rcon"

game="${1-}"
shift || true

if [ -z "$game" ]; then
	echo 'Error: game not specified.' >&2
	help
	exit 1
fi

game_cfg_dir="$source_dir/cfg/$game"

if [ ! -d "$game_cfg_dir" ]; then
	echo "Error: unknown game: $game" >&2
	exit 1
fi

if [ "${1-}" = '-r' ]; then
	root=true
	shift
fi

root=${root-false}

# read user from config.json:
# {
#   "user": {
#     "name": "server-user:30120",
#     "group": "server-group:30120"
#   }
# }
config() {
	key="$1"
	jq --raw-output ".${key}" "$game_cfg_dir/config.json"
}

user_namestr="$(config 'user.name')"
user_groupstr="$(config 'user.group')"

user_name="${user_namestr%%:*}"
user_group="${user_groupstr%%:*}"
user_uid="${user_namestr##*:}"
user_gid="${user_groupstr##*:}"

host_hot="$source_dir/server-files/$game/hot"

mounts=(
	"$source_dir:/src"
	"$host_hot:/hot"
)

print_mounts() {
	local mounts=("$@")

	if [ -z "$mounts" ]; then
		return
	fi

	echo "Mounts:"
	for m in "${mounts[@]}"; do
		host="${m%%:*}"
		guest="${m#*:}"
		echo "$guest -> $host"
	done
}

preamble() {
	cat <<-EOF

		Game: $game

		$(print_mounts "${mounts[@]}")

		EOF

	column \
		--table \
		--separator : \
		--output-separator : \
		--table-right 1 \
	<<-EOF
		$user_name:$user_group
		 $user_uid:$user_gid
	EOF
	echo
}
preamble

("$source_dir/m1/py.sh" -- -m cli "$game" download)

run_mounts=()
for m in "${mounts[@]}"; do
	run_mounts+=(--volume "$m")
done

docker_run=(
	docker run --rm -it
	--name "$game-shell"
	--hostname "$game"
	--workdir '/hot'
	"${run_mounts[@]}"
)

$root && docker_run+=(--user root)

"${docker_run[@]}" "$game"
