#!/bin/bash
set -euo pipefail

# m1/py.sh -- -m cli vrising start

hot_dir="$1"
rcon_password="$2"

# TODO: Get from host/config?
game_port=34480

cd "$hot_dir" || exit 1

# https://stackoverflow.com/questions/12050021/how-to-make-xvfb-display-visible

if ! pgrep x11vnc; then
	x11vnc \
		-bg \
		-display WAIT$DISPLAY \
		-forever \
		-nopw \
		-quiet \
		-xkb &
	Xvfb \
		$DISPLAY \
		-screen 0 640x480x16 \
		-nolisten tcp &
fi

server_cfg="$hot_dir/Icarus/Saved/Config/WindowsServer/ServerSettings.ini"
if [ -f "$server_cfg" ]; then
	# https://github.com/RocketWerkz/IcarusDedicatedServer/wiki/Server-Config-&-Launch-Parameters#command-line-args
	sed -i \
		-e "s/^AdminPassword=.*$/AdminPassword=$rcon_password/" \
		-e "s/^ServerName=.*$/ServerName=Nerdhaven/" \
		-e "s/^JoinPassword=.*$/JoinPassword=121212/" \
		"$server_cfg"
fi

# https://github.com/RocketWerkz/IcarusDedicatedServer
wine IcarusServer.exe \
	-SteamServerName="Nerdhaven" \
	-PORT="$game_port" \
	-ResumeProspect

if [ ! -f /tmp/stopfile ]; then
	# If there is no stopfile, wait to prevent rapid restart loop if the server
	# is failing to start.
	sleep 10
fi
