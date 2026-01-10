#!/bin/bash
set -euo pipefail

# m1/py.sh -- -m cli vrising start

hot_dir="$1"
rcon_password="$2"

# TODO: Get from host/config?
game_port=34360
rcon_port=34111

cd "$hot_dir" || exit 1

export DISPLAY=:1
Xvfb "$DISPLAY" -screen 0 640x480x16 -nolisten tcp -nolisten unix &

export WINEPREFIX="$(pwd)/wine" # TODO: Share from host? This dir is huge.

cfg_dir="$hot_dir/save-data/Settings"
host_cfg="$cfg_dir/ServerHostSettings.json"
game_cfg="$cfg_dir/ServerGameSettings.json"

if [ ! -f "$host_cfg" ]; then
	tar -C "$hot_dir/VRisingServer_Data/StreamingAssets/Settings" -cf - . |
		tar -C "$cfg_dir" -xf -
fi

tmp_cfg="$(mktemp)"
jq "$(
	cat <<-EOF
		.Rcon.Enabled = true |
		.Password = "vampirebtw" |
		.Rcon.Port = $rcon_port |
		.Rcon.Password = "$rcon_password" |
		.Port = $game_port
	EOF
)" "$host_cfg" >"$tmp_cfg"
cat "$tmp_cfg" >"$host_cfg"
rm "$tmp_cfg"

tmp_cfg="$(mktemp)"
jq "$(
	cat <<-EOF
		.GameModeType = "PvE" |
		.ClanSize = 20 |
		.TeleportBoundItems = false
	EOF
)" "$game_cfg" >"$tmp_cfg"
cat "$tmp_cfg" >"$game_cfg"
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
