#!/bin/bash
set -euo pipefail

hot_dir="$1"

steamcmd \
	+force_install_dir "$hot_dir" \
	+login anonymous \
	+app_update 1169370 validate \
	+quit
