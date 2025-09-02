#!/bin/bash
set -euo pipefail

# m1/py.sh -- -m cli vrising start

hot_dir="$1"
rcon_password="$2"

touch /tmp/stopfile  # TODO: Remove this line

game_port=34360

cd "$hot_dir" || exit 1

Xvfb :1 -screen 0 640x480x16 -nolisten tcp -nolisten unix &
export DISPLAY=:1

export WINEPREFIX="$(pwd)/wine"

cfg_dir="$hot_dir/save-data/Settings"
host_cfg="$cfg_dir/ServerHostSettings.json"

if [ ! -f "$host_cfg" ]; then
	tar -C "$hot_dir/VRisingServer_Data/StreamingAssets/Settings" -cf - . | \
		tar -C "$cfg_dir" -xf -
fi

tmp_cfg="$(mktemp)"

jq "$(cat <<-EOF
	.Rcon.Enabled = true |
	.Password = "vampirebtw" |
	.Rcon.Port = 34010 |
	.Rcon.Password = "$rcon_password" |
	.Port = $game_port
	EOF
)" "$host_cfg" > "$tmp_cfg"

cat "$tmp_cfg" > "$host_cfg"
rm "$tmp_cfg"

# https://vrising.fandom.com/wiki/V_Rising_Dedicated_Server#Launch_Commands
wine VRisingServer.exe \
	-persistentDataPath '.\save-data'

if [ ! -f /tmp/stopfile ]; then
	# If there is no stopfile, wait to prevent rapid restart loop if the server
	# is failing to start.
	# TODO: Move this to manage?
	sleep 10
fi
