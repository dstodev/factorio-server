#!/bin/bash
set -euo pipefail

hot_dir="$1"

# V Rising only supports Windows dedicated server as of: Sep. 1, 2025

steamcmd \
	+@sSteamCmdForcePlatformType windows \
	+login anonymous \
	+force_install_dir "$hot_dir" \
	+app_update 1829350 validate \
	+quit
