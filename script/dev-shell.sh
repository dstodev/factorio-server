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

		-h, --help      Print this message.
		-v, --verbose   Enable verbose output.
		-r, --root      Run as root instead of as server's user.
	EOF
}

status=0
getopt --test >/dev/null 2>&1 || status=$?
if [ "$status" -ne 4 ]; then
	echo 'Error: getopt --test failed. This script requires GNU getopt from util-linux.' >&2
	exit 1
fi

status=0
args=$(getopt --name "$(basename -- "$0")" \
	--options hrv \
	--longoptions help,verbose,root \
	-- "$@") || status=$?

if [ "$status" -ne 0 ]; then
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
	-v | --verbose)
		verbose=true
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

verbose="${verbose-false}"
root="${root-false}"

if $verbose; then
	printf -- 'Args: |%s|\n' "$args"
fi

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
repo_dir="$(readlink -f -- "$this_dir/..")"

game="${1-}"
shift || true

if [ -z "$game" ]; then
	echo 'Error: game not specified.' >&2
	help
	exit 1
fi

game_cfg_dir="$repo_dir/cfg/$game"

if [ ! -d "$game_cfg_dir" ]; then
	echo "Error: unknown game: $game" >&2
	exit 1
fi

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

host_hot="$repo_dir/server-files/$game/hot"
host_shelf="$repo_dir/shelf/$game"

mounts=(
	"$repo_dir:/src"
	"$host_hot:/hot"
	"$host_shelf:/shelf"
)

preamble() {
	echo
	echo 'Mounts:'
	column \
		--table \
		--separator '>' \
		--output-separator '>' \
		--table-right 1 \
		<<-EOF
			$(print_mounts "${mounts[@]}")
		EOF
	echo
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

print_mounts() {
	local mounts=("$@")

	if [ -z "$mounts" ]; then
		return
	fi

	for m in "${mounts[@]}"; do
		host="${m%%:*}"
		guest="${m#*:}"
		echo "$guest -> $host"
	done
}

preamble

verbose_flag=
if $verbose; then
	verbose_flag='--verbose'
fi

("$repo_dir/m1/py.sh" -- -m cli $verbose_flag "$game" download) # builds the image

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
