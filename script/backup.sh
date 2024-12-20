#!/bin/bash
set -euo pipefail

# Add to crontab, running every hour except 5am (time of server restart):
#  crontab -e
#  0 0-4,6-23 * * * /absolute/path/to/script/backup.sh >/dev/null 2>&1

case ${1-} in
-f | --force) force=true ;;
esac

force=${force-false}

script_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink -f "$script_dir/..")"

backup_dir="$source_dir/backups"

mkdir --parents --verbose "$backup_dir"

rcon="$script_dir/send-rcon.sh"

save_cmd='/server-save'

if $force; then
	# Try to save, but continue on error.
	"$rcon" "$save_cmd" || true
else
	# Exit on error (from 'set -e' above)
	# e.g. server is not running (and thus does not respond to RCON)
	# server does not need to be running to save, but this prevents
	# automated backups when the server is not running.
	"$rcon" "$save_cmd"
fi

server_data_path="$source_dir/server-files/factorio/saves"

# shellcheck source=script/util.sh
source "$script_dir/util.sh"

backup_target="$backup_dir/$(timestamp).tar"

# Copy files to a temporary directory before archiving.
tmp_dir="$script_dir/tmp-backup"
mkdir --parents --verbose "$tmp_dir"
rsync --archive --no-compress --delete --exclude '*.tmp*' "$server_data_path" "$tmp_dir"
tar --create --file "$backup_target" --directory "$tmp_dir" '.'
echo "Backup saved to: $backup_target"
rm --recursive --verbose "$tmp_dir" | tail --lines 1

# Preserve most recent backups, deleting older ones.
days_to_preserve=14
find "$backup_dir" -maxdepth 1 -type f -mtime +$days_to_preserve -name '*.tar' -delete
