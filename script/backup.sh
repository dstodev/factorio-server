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

save_cmd='/server-save'

if $force; then
	# Try to save, but continue on error.
	"$rcon" "$save_cmd" >/dev/null 2>&1 || true
else
	# Exit on error (from 'set -e' above)
	# e.g. server is not running (and thus does not respond to RCON)
	# server does not need to be running to save, but this prevents
	# automated backups when the server is not running.
	"$rcon" "$save_cmd"
fi

server_data_path="$source_dir/server-files/factorio/saves"

timestamp=$(date +"%Y%m%dT%H%M%S%z")
# Based on ISO 8601:
# - https://en.wikipedia.org/wiki/ISO_8601
# - https://www.gnu.org/software/coreutils/manual/html_node/Options-for-date.html#index-_002dI_005btimespec_005d
# Try:
#   date --iso-8601=seconds
# or
#   date +"%Y-%m-%dT%H:%M:%S%:z"

backup_target="$backup_dir/$timestamp.tar"

# Copy files to a temporary directory before archiving.
tmp_dir="$script_dir/tmp"
rsync --archive --no-compress --delete --exclude '*.tmp*' "$server_data_path" "$tmp_dir"
tar --create --file "$backup_target" --directory "$tmp_dir" . # note dot at end
rm --recursive --force "$tmp_dir"

# Preserve most recent backups, deleting older ones.
days_to_preserve=14
find "$backup_dir" -maxdepth 1 -type f -mtime +$days_to_preserve -name '*.tar' -delete
