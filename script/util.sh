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
	#   date +'%Y-%m-%dT%H:%M:%S%:z' # can use %z instead of %:z
	#   date +'%Y-%m-%dT%H:%M:%SZ' --utc
	#
	date +"%Y-%m-%dT%H+%M+%S%z"
}

# Print '-q' if VERBOSE is unset or 0.
# Useful to "forward" non-verbosity to commands supporting -q.
flag_quiet() {
	quiet '-q'
}

# Print '-v' if VERBOSE is set and greater than 0.
# Useful to "forward" verbosity to commands supporting -v.
flag_verbose() {
	verbose '-v'
}

# Print a message if VERBOSE is unset or 0.
#
# Use carefully: printing a message is conventionally antithetical to "quiet"
# operation.
quiet() {
	if [ "${VERBOSE:-0}" -eq 0 ]; then
		echo "$@"
	fi
}

# Print a message if VERBOSE is set and greater than 0.
verbose() {
	if [ "${VERBOSE:-0}" -gt 0 ]; then
		echo "$@"
	fi
}

# Find and return a working python interpreter of at least the provided version
#
# Usage: py_interpreter <min_version>
#
# Example:
#   > py="$(py_interpreter 3.10)"
#   > "$py" --version
#   Python 3.10.12
py_interpreter() {
	version="${1-3.0.0}"
	IFS=. read -r major minor patch <<<"$version"

	major="${major:-0}"
	minor="${minor:-0}"
	patch="${patch:-0}"

	version="${major}.${minor}.${patch}"

	command_candidates=(
		python3
		python
	)

	for candidate in "${command_candidates[@]}"; do
		candidate_path="$(command -v "$candidate")" || true
		if [ -z "$candidate_path" ]; then
			verbose "-- not found: $candidate" >&2
			continue
		fi

		candidate_version="$("$candidate" --version 2>&1 | grep -o '[0-9]\+\.[0-9]\+\.[0-9]\+')"
		if version_ge "$candidate_version" "$version"; then
			echo "$candidate_path"
			return 0
		else
			verbose "-- rejected: $candidate_path ($candidate_version)" >&2
		fi
	done

	echo "Error: No suitable Python interpreter!" >&2
	echo "       Please install $version or higher." >&2

	return 1
}

# True if version $1 >= $2
version_ge() {
	# Sorts in ascending order, so the first line is the lowest
	# If $2 is the lowest version, then $1 >= $2
	[ "$(printf '%s\n' "$1" "$2" | sort --version-sort | head --lines 1)" = "$2" ]
}
