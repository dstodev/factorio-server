#!/bin/bash
set -euo pipefail

hot_dir="$1"

# https://steamdb.info/app/2089300/depots/
app_id=2089300

# Dedicated server only supports Windows as of: Oct. 5, 2025
steamcmd \
	+@sSteamCmdForcePlatformType windows \
	+login anonymous \
	+force_install_dir "$hot_dir" \
	+app_update "$app_id" validate \
	+quit
