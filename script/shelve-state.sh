#!/bin/bash
set -euo pipefail

script_dir="$(builtin cd -- "$(dirname "$0")" && pwd -P)"
source_dir="$(readlink --canonicalize "$script_dir/..")"

shelf_dir="$source_dir/shelf"

help() {
	cat <<-EOF
		Usage: $(basename "$0") [ -l ]
		  -h, --help  Print this message.
		  -l, --print-latest
		              Print absolute path to latest-shelved server files.
	EOF
}

canonical=$(getopt --name "$(basename "$0")" \
	--options hl \
	--longoptions help,print-latest \
	-- "$@") || status=$?

if [ "${status-0}" -ne 0 ]; then
	help
	exit 1
fi

eval set -- "$canonical"

while :; do
	case "$1" in
	-h | --help)
		help
		exit 0
		;;
	-l | --print-latest)
		print_latest=true
		;;
	--)
		shift # --
		break
		;;
	esac
	shift # option
done

print_latest=${print_latest-false}

latest="$(
	find "$shelf_dir" -mindepth 1 -maxdepth 1 -type d -printf '%f\0' 2>/dev/null |
		sort --zero-terminated --numeric-sort |
		tail --zero-terminated --lines 1 |
		xargs --null
)" || true

if $print_latest; then
	latest_dir="$shelf_dir/$latest"
	[ -n "$latest" ] && echo "$latest_dir" && exit 0 || exit 1
fi

shelf_targets=(
	# "$source_dir/backups"
	"$source_dir/logs"
	"$source_dir/server-files"
)

check() {
	[ -e "$1" ]
}

verified=()

for target in "${shelf_targets[@]}"; do
	check "$target" && verified+=("$target")
done

next=$((latest + 1))
next_dir="$shelf_dir/$next"

if [ "${#verified[@]}" -gt 0 ]; then
	mkdir --parents "$next_dir"
	mv --verbose --target-directory="$next_dir" "${verified[@]}"
fi
