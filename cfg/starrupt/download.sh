#!/bin/bash
set -euo pipefail

hot_dir="$1"

# Star Rupture dedicated server only supports Windows as of: Jan 9, 2026
# https://steamdb.info/app/3809400/depots/
steamcmd \
	+@sSteamCmdForcePlatformType windows \
	+force_install_dir "$hot_dir" \
	+login anonymous \
	+app_update 3809400 validate \
	+quit
