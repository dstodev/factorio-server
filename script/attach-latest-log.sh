#!/bin/bash
set -euo pipefail

case ${1-} in
-f | --follow) follow=true ;;
esac

follow=${follow-false}

script_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink --canonicalize "$script_dir/..")"

logs_dir="$source_dir/logs"

if logs=$(ls -At "$logs_dir" 2>/dev/null) && [ -n "$logs" ]; then
	newest_log=$(echo "$logs" | head -n 1)
	cmd=(less)
	[ "$follow" = true ] && cmd+=(+F)
	"${cmd[@]}" "$logs_dir/$newest_log"
else
	echo 'No log files found!'
fi
