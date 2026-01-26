#!/bin/bash
set -euo pipefail

this_dir="$(dirname -- "$(readlink -f -- "$0")")"
repo_dir="$(readlink -f -- "$this_dir/../..")"
log_root="$repo_dir/logs"

game="$(basename -- "$this_dir")"
log_dir="$log_root/$game"

# Log files have format '%Y-%m-%d_%H-%M-%SZ', e.g.:
# 	logs/atm10/2025-12-04_00-34-10Z.log
# 	(from: m1/src/manage/util.py)

latest_log() {
	local dir="$1"
	find "$dir" -type f -printf '%T+ %p\0' |
		LC_ALL=C sort --zero-terminated --reverse |         # logs sorted by modify time, newest first
		head --zero-terminated --lines=1 |                  # newest (first) log only
		cut --zero-terminated --delimiter=' ' --fields=2- | # filename only
		xargs -0 echo                                       # replace null with newline
}

# Lines in chat look like:
# 	[01:29:53] [Server thread/INFO] [minecraft/MinecraftServer]: <Nynja> oh no

egrep='grep --extended-regexp --line-buffered'
s='[[:space:]]'

log_file="$(latest_log "$log_dir")"

# follow the newest log
tail --follow=descriptor --lines=+0 "$log_file" |
	$egrep "^(\[[^]]+\]$s*){3}:$s<[^>]+>.*$" |
	$egrep --only-matching '<[^>]+>.*$'
