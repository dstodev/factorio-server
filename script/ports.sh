#!/bin/bash
set -euo pipefail

# Find all ../cfg/*/config.json and print configured ports for each

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
server_dir="$(readlink -f -- "$this_dir/..")"
cfg_dir="$server_dir/cfg"

declare -A ports

while IFS= read -r -d $'\0' cfg_file; do
	server_name="$(basename "$(dirname "$cfg_file")")"
	port="$(jq -r '.port.game' "$cfg_file")"
	ports[$port]+=$'\n'"$server_name"
done < <(find "$cfg_dir" -mindepth 2 -maxdepth 2 -name 'config.json' -print0)

for port in $(printf '%s\0' "${!ports[@]}" |
	sort --numeric-sort --zero-terminated |
	xargs --null); do
	games="$(sed '/^$/d' <<<"${ports[$port]}" | paste -sd, -)"
	printf '%s:%s\n' "$port" "$games"
done

# TODO: Transpose using something like:
#
# paste <(cmd1) <(cmd2) <(cmd3)
#
# subs=()
# for cmd in "${cmds[@]}"; do
#     subs+=( <(eval "$cmd") )
# done
# paste "${subs[@]}"
