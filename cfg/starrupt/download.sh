#!/bin/bash
set -euo pipefail

hot_dir="$1"

# https://steamdb.info/app/3809400/depots/
app_id=3809400

retries=0
max_retries=5

while :; do
	status=0

	# Dedicated server only supports Windows as of: Jan 9, 2026
	steamcmd \
		+@sSteamCmdForcePlatformType windows \
		+force_install_dir "$hot_dir" \
		+login anonymous \
		+app_update "$app_id" validate \
		+quit || status=$?
	if [ "$status" -eq 0 ]; then
		break
	else
		echo "!! steamcmd failed with code: $status" >&2
		retries=$((retries + 1))
		if [ "$retries" -gt "$max_retries" ]; then
			echo "!! Reached maximum retries ($max_retries). Exiting." >&2
			exit 1
		fi
		sleep 5
	fi
done
