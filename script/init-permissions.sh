#!/bin/bash
set -euo pipefail

# This script configures the server user and group, and sets
# permissions on the server files.

if ! sudo --non-interactive true 2>/dev/null; then
	# If not yet sudo, print prompt and read password before loop
	printf 'sudo password required to fix server file permissions.\n'
	sudo --validate || exit
fi

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
source_dir="$(readlink --canonicalize "$script_dir/..")"

docker_dir="$source_dir/docker"
server_dir="$source_dir/server-files"

set -o allexport
# shellcheck source=docker/.env
source "$docker_dir/.env"
set +o allexport

server_group_id="$SERVER_GROUP_ID"
server_group_name="$SERVER_GROUP_NAME"

# If there is no server group, create it.
if [ ! "$(getent group "$server_group_name")" ]; then
	echo Creating group: "$server_group_name"
	sudo groupadd --gid "$server_group_id" "$server_group_name"
fi

server_user_id="$SERVER_USER_ID"
server_user_name="$SERVER_USER_NAME"

# If there is no server user, create it.
if [ ! "$(getent passwd "$server_user_name")" ]; then
	echo Creating user: "$server_user_name"
	sudo useradd --uid "$server_user_id" --gid "$server_group_id" "$server_user_name"
fi

# If the host user is not in the server group, add them.
if ! id --groups --name | grep --quiet --fixed-strings --word-regexp "$server_group_name"; then
	echo "Adding current user $(id --user --name) to group: $server_group_name"
	sudo usermod -aG "$server_group_name" "$(id --user --name)"
	echo 'Log-out and back-in to apply group changes.'
fi

printf 'Setting host permissions: '
sudo chown --recursive ":$server_group_name" "$source_dir"
sudo find "$source_dir" -type d -exec chmod g+w,g+s {} +
sudo chmod g+w,g+x "$source_dir/cfg/start.sh"
find "$source_dir" -maxdepth 0 -printf '%p => [%M] %u:%g\n'

mkdir --parents "$server_dir"

printf 'Setting server permissions: '
sudo chown --recursive "$server_user_name:$server_group_name" "$server_dir"
sudo find "$server_dir" -type d -exec chmod g+w,g+s {} +
find "$server_dir" -maxdepth 0 -printf '%p => [%M] %u:%g\n'
