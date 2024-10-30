#!/bin/bash
set -euo pipefail

# Add to crontab, running every hour except 5am (time of server restart):
#  crontab -e
#  0 0-4,6-23 * * * /absolute/path/to/script/backup.sh >/dev/null 2>&1

case ${1-} in
-f | --force) force=true ;;
esac

force=${force-false}

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
source_dir="$(readlink --canonicalize "$script_dir/..")"

backup_dir="$source_dir/backups"

mkdir --parents "$backup_dir"

rcon="$script_dir/send-rcon.sh"

if $force; then
	# Try to save, but continue on error.
	"$rcon" "/server-save" >/dev/null 2>&1 || true
else
	# Exit on error (from 'set -e' above)
	# e.g. server is not running (and thus does not respond to RCON)
	# server does not need to be running to save, but this prevents
	# automated backups when the server is not running.
	"$rcon" "/server-save"
fi

compose_yml="$source_dir/docker/compose.yml"
compose=(docker compose --file "$compose_yml")

"${compose[@]}" run --rm backup

# Preserve most recent backups, deleting older ones.
days_to_preserve=14
find "$backup_dir" -maxdepth 1 -type f -mtime +$days_to_preserve -name '*.tar.bz2' -delete
