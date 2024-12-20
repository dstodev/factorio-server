#!/bin/bash
set -euo pipefail

# If build-essential is not installed, prompt to install it.
if ! dpkg --status build-essential >/dev/null 2>&1; then
	printf '%s\n%s\n\n%s' \
		'!! build-essential is not installed.' \
		'!! This is required to build the RCON client.' \
		'Install build-essential? (y/N): '

	read -r option

	case $option in
	[Yy]*) sudo apt update && sudo apt install --yes build-essential ;;
	*)
		echo 'Operation aborted.'
		exit 1
		;;
	esac
fi

script_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink --canonicalize "$script_dir/..")"

docker_dir="$source_dir/docker"
rcon_dir="$source_dir/rcon"

set -o allexport
# shellcheck source=docker/.env
source "$docker_dir/.env"
set +o allexport

host_str="localhost:$RCON_PORT"

source_path="$rcon_dir/main.cxx"
client_path="$rcon_dir/client.out"
secret_path="$rcon_dir/secret"

# If the RCON client does not exist, or the source file is newer than the
# client, (re)build the client.
if [ ! -f "$client_path" ] || [ "$source_path" -nt "$client_path" ]; then
	printf 'Building RCON client... '
	g++ -O3 -o "$client_path" "$source_path"
	printf 'done!\n'
	"$client_path" test
fi

# If the RCON secret file does not exist, tell the user to create it.
if [ ! -f "$rcon_dir/secret" ]; then
	echo '!! RCON password file does not exist.'
	echo "Create '$secret_path' containing only the RCON password."
	exit 1
fi

"$client_path" "$host_str" "$@" <"$secret_path" | sed 's/^/(rcon) /'
