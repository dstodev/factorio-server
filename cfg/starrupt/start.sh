#!/bin/bash
set -euo pipefail

# if [ "$(id --user)" -eq 0 ]; then
# 	echo "! Starting dbus-daemon"
# 	mkdir -p /run/dbus
# 	dbus-daemon --system --fork
# 	su server-user -c "$0" "$@"
# fi

hot_dir="$1"
rcon_password="$2"

# TODO: Get from host/config?
game_port=7777

WEBVIEW_RESOLUTION='1280x720x24'
WEBVIEW_PORT_HTTP_INTERNAL=8080
WEBVIEW_PORT_X11VNC_INTERNAL=5900

export DBUS_SYSTEM_BUS_ADDRESS='unix:path=/run/dbus/system_bus_socket'

eval "$(dbus-launch --sh-syntax)"
export DBUS_SESSION_BUS_ADDRESS
export DBUS_SESSION_BUS_PID

BG_SERVICE_PIDS=()
start_bg_service() {
	local cmd="$1"
	local pid
	shift
	local log_dir="$HOME/service-logs"
	local log_pfx="$log_dir/$cmd"
	mkdir -p "$log_dir"
	if ! pgrep --exact "$cmd" >/dev/null; then
		echo "$cmd $*"
		"$cmd" "$@" >"$log_pfx.out.log" 2>"$log_pfx.err.log" &
		pid=$!
		ps -p "$pid"
		BG_SERVICE_PIDS+=("$pid")
	fi
}

# Create virtual framebuffer
start_bg_service Xvfb "$DISPLAY" \
	-nolisten tcp \
	-screen 0 "$WEBVIEW_RESOLUTION"
# -nolisten unix

# Start X11 window manager
start_bg_service openbox

# Start VNC server to share X11 display
start_bg_service x11vnc \
	-display "WAIT$DISPLAY" \
	-forever \
	-listen localhost \
	-nopw \
	-rfbport "$WEBVIEW_PORT_X11VNC_INTERNAL"
#	removed `-ncache 10` because it causes weird screen duplication

# Start websockify to provide noVNC access to VNC server over HTTP
start_bg_service websockify \
	--web=/usr/share/novnc/ \
	"$WEBVIEW_PORT_HTTP_INTERNAL" "localhost:$WEBVIEW_PORT_X11VNC_INTERNAL"

# Set up environment
# TODO: Mount with Docker to host
#mkdir -p "$HOME/vault"

# Set "desktop" background color
xsetroot -solid gray

# for svc_pid in "${BG_SERVICE_PIDS[@]}"; do
# 	echo "- Awaiting process exit: $svc_pid"
# 	wait "$svc_pid"
# done

cd "$hot_dir"

wan_ip="$(curl https://checkip.amazonaws.com)"

# https://www.survivalservers.com/wiki/index.php?title=How_to_Create_a_StarRupture_Server_Guide
wine StarRuptureServerEOS.exe \
	-Log \
	-ServerName="Nerdhaven" \
	-Port="$game_port" \
	-QueryPort=27015 \
	-MULTIHOME="$wan_ip"

if [ ! -f /tmp/stopfile ]; then
	# If there is no stopfile, wait to prevent rapid restart loop if the server
	# is failing to start.
	sleep 10
fi
