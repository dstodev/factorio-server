#!/bin/bash
set -euo pipefail

this_dir="$(dirname -- "$(readlink -f -- "$0")")"

# shellcheck source=docker/util/docker.sh
. "$this_dir/docker.sh"

lan_ip="$(ip route get 1 | awk '{ print $7; exit }')"
(
	compose_env
	printf '%s:%s\n' "$lan_ip" "$GW_PORT_ADMIN"
)
