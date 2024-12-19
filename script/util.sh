# This file is meant to be sourced, not executed.
# shellcheck shell=bash

timestamp() {
	# Format based on ISO 8601:
	# - https://en.wikipedia.org/wiki/ISO_8601
	# - https://www.gnu.org/software/coreutils/manual/html_node/Options-for-date.html#index-_002dI_005btimespec_005d
	#
	# with some characters replaced for portability with other operating systems
	# which do not support, among others, the ':' character in paths (e.g. Windows).
	#
	# Try commands:
	#   date --iso-8601=seconds
	# or:
	#   date +"%Y-%m-%dT%H:%M:%S%:z"
	date +"%Y-%m-%dT%H+%M+%S%z"
}
