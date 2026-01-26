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

cleanup() {
	rm -rfv Astro/Saved/EXITREQUEST
}
cleanup

death_loop_delay() {
	if [ ! -f /tmp/stopfile ]; then
		local delay=10
		# If there is no stopfile, wait to prevent rapid restart loop if
		# the server is failing to start.
		cat <<-EOF >&2
			!! Server stopped unexpectedly!
			!! Waiting $delay seconds before restart...
		EOF
		sleep $delay
	fi
}
on_exit() {
	cleanup
	set +x
	death_loop_delay
}
trap on_exit EXIT

# TODO: Get from host/config?
game_port=58777

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

# Set "desktop" background color
xsetroot -solid gray

# for svc_pid in "${BG_SERVICE_PIDS[@]}"; do
# 	echo "- Awaiting process exit: $svc_pid"
# 	wait "$svc_pid"
# done

#redist_dir="$hot_dir/_CommonRedist"

# set -x
# wine "$redist_dir/vcredist/2015/vc_redist.x86.exe" /quiet /norestart
# wine "$redist_dir/vcredist/2015/vc_redist.x64.exe" /quiet /norestart
# wine "$redist_dir/DirectX/Jun2010/DXSETUP.exe" /silent

# wine "$hot_dir/Engine/Extras/Redist/en-us/UE4PrereqSetup_x64.exe" /s
# set +x

wan_ip="$(curl https://checkip.amazonaws.com)"

printf -- '-- Address: %s:%s\n' "$wan_ip" "$game_port"

cfg_dir="$hot_dir/Astro/Saved/Config/WindowsServer"

if [ -f "$cfg_dir/Engine.ini" ]; then
	if ! grep -q "\[URL\]" "$cfg_dir/Engine.ini"; then
		cat <<-EOF >>"$cfg_dir/Engine.ini"
			[URL]
			Port=
		EOF
	fi
	sed -i "s/^Port=.*$/Port=$game_port/" "$cfg_dir/Engine.ini"

	sed -i "s/^PublicIP=.*$/PublicIP=$wan_ip/" "$cfg_dir/AstroServerSettings.ini"
	sed -i "s/^ServerName=.*$/ServerName=Nerdhaven/" "$cfg_dir/AstroServerSettings.ini"
	sed -i "s/^ServerPassword=.*$/ServerPassword=121212/" "$cfg_dir/AstroServerSettings.ini"
	sed -i "s/^ConsolePassword=.*$/ConsolePassword=$rcon_password/" "$cfg_dir/AstroServerSettings.ini"
fi

sync

cd "$hot_dir"

set -x

# https://blog.astroneer.space/p/astroneer-dedicated-server-details/
wine "Astro/Binaries/Win64/AstroServer-Win64-Shipping.exe"
