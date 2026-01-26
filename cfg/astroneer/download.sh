#!/bin/bash
set -euo pipefail

hot_dir="$1"

retries=0
max_retries=1

while :; do
	status=0

	# Dedicated server only supports Windows as of: Jan 13, 2026
	# https://blog.astroneer.space/p/astroneer-dedicated-server-details/
	# https://steamdb.info/app/728470/depots/
	steamcmd \
		+@sSteamCmdForcePlatformType windows \
		+force_install_dir "$hot_dir" \
		+login anonymous \
		+app_update 728470 validate \
		+app_update 1007 validate \
		+app_update 228980 validate \
		+quit || status=$?
	if [ "$status" -eq 0 ]; then
		break
	else
		echo "steamcmd failed with code: $status" >&2
		retries=$((retries + 1))
		if [ "$retries" -gt "$max_retries" ]; then
			echo "Reached maximum retries ($max_retries). Exiting." >&2
			exit 1
		fi
		sleep 5
	fi
done
