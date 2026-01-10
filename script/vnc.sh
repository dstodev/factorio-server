#!/bin/bash
set -euo pipefail

# Assuming the game container has a vnc server running, start a container
# connected to its network to expose the VNC server port to the host.

if [ $# -eq 0 ]; then
	echo "Usage: $0 <game>" >&2
	exit 1
fi

game="$1"

ctr="$game-server"

port_host=8080  # Host port to expose VNC on
port_guest=8080 # Guest (container) port where VNC server is listening

network=$(docker inspect "$ctr" --format '{{range .NetworkSettings.Networks}}{{.NetworkID}}{{end}}' | head -n 1)
ctr_ip=$(docker inspect "$ctr" --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' | head -n 1)

cat <<-EOF
	Relaying 0.0.0.0:$port_host to $ctr ($ctr_ip):$port_guest
EOF

docker run --rm \
	--name vnc \
	--network $network \
	--publish $port_host:$port_guest \
	alpine/socat \
	TCP-LISTEN:$port_guest,fork,reuseaddr,nodelay TCP-CONNECT:$ctr_ip:$port_guest,nodelay
