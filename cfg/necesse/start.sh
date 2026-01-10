#!/bin/bash
set -euo pipefail

# m1/py.sh -- -m cli vrising start

hot_dir="$1"
rcon_password="$2"

cd "$hot_dir" || exit 1

mkdir -p save
mkdir -p ~/.config
ln -s "$hot_dir/save" ~/.config/Necesse

# https://necessewiki.com/Multiplayer-Linux
./StartServer-nogui.sh -world 'necesseworld'

if [ ! -f /tmp/stopfile ]; then
	# If there is no stopfile, wait to prevent rapid restart loop if the server
	# is failing to start.
	sleep 10
fi
