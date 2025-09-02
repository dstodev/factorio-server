#!/bin/bash
set -euo pipefail

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink -f -- "$this_dir/../..")"
rcon_dir="$source_dir/rcon"

game="$(basename -- "$this_dir")"

# read ports and user from server.json:
# {
#   "port": {
#     "game": 34120,
#     "rcon": 34240
#   },
#   "user": {
#     "name": "server-user:30120",
#     "group": "server-group:30120"
#   }
# }
config() {
	key="$1"
	jq --raw-output ".${key}" "$this_dir/server.json"
}

game_port="$(config 'port.game')"
user_namestr="$(config 'user.name')"
user_groupstr="$(config 'user.group')"

user_name="${user_namestr%%:*}"
user_group="${user_groupstr%%:*}"
user_uid="${user_namestr##*:}"
user_gid="${user_groupstr##*:}"

cat <<-EOF
	user  name: $user_name
	       uid: $user_uid
	     group: $user_group
	       gid: $user_gid

	port: $game_port
	EOF

("$source_dir/m1/py.sh" -vp -- -m cli "$game" download)

hot_dir="$source_dir/server-files/$game/hot"

docker run --rm -it \
	-v "$hot_dir:/hot" \
	-v "$source_dir:/src:ro" \
	"$game"

# -p "$game_port:$game_port/udp" \
