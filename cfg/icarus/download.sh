#!/bin/bash
set -euo pipefail

hot_dir="$1"

# Icarus dedicated server only supports Windows as of: Oct. 5, 2025

steamcmd \
	+@sSteamCmdForcePlatformType windows \
	+login anonymous \
	+force_install_dir "$hot_dir" \
	+app_update 2089300 validate \
	+quit
