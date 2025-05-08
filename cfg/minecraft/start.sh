#!/bin/bash
set -euo pipefail

server_dir="$1"
rcon_password="$2"

touch /tmp/stopfile

game_port=34120
rcon_port=34240

cd "$server_dir" || exit 1

mapfile -t properties_files < <(find . -name "server.properties" -type f)

for properties_file in "${properties_files[@]}"; do
	echo "Patching: $properties_file"

	# Critical settings
	sed -i "s/^enable-rcon=false/enable-rcon=true/" "$properties_file"
	sed -i "s/^rcon.password=.*/rcon.password=$rcon_password/" "$properties_file"
	sed -i "s/^rcon.port=.*/rcon.port=$rcon_port/" "$properties_file"
	sed -i "s/^server-port=.*/server-port=$game_port/" "$properties_file"

	# Optional settings
	sed -i "s/^allow-flight=false/allow-flight=true/" "$properties_file"
	sed -i "s/^spawn-protection=.*/spawn-protection=0/" "$properties_file"
done

"$server_dir/start-server.sh"
