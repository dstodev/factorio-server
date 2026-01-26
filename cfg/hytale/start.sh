#!/bin/bash
set -euo pipefail

hot_dir="$1"
rcon_password="$2"

death_loop_delay() {
	if [ ! -f /tmp/stopfile ]; then
		# If there is no stopfile, wait to prevent rapid restart loop if
		# the server is failing to start.
		sleep 10
	fi
}
cleanup() {
	death_loop_delay
}
trap cleanup EXIT

# TODO: Get from host/config?
game_port=55520

#wan_ip="$(curl https://checkip.amazonaws.com)"

cd "$hot_dir/Server"

if [ -f config.json ]; then
	sed -i 's|"ServerName":.*,|"ServerName": "Nerdhaven",|' config.json
	sed -i 's|"MOTD":.*,|"MOTD": "Welcome!",|' config.json
	sed -i 's|"Password":.*,|"Password": "121212",|' config.json
	sed -i 's|"World":.*,|"World": "Nerdhaven",|' config.json
fi

java -XX:AOTCache=HytaleServer.aot \
	-jar HytaleServer.jar \
	--assets "$hot_dir/Assets.zip" \
	--bind "$game_port"
