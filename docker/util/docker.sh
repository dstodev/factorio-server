# shellcheck shell=bash

# :: compose.sh ------------------------------------
#
# Source this file with Bash for Docker helpers.
#
# --------------------------------------------------

# Run `docker compose "$@"` within directory: docker/
compose() {
	local this_dir repo_dir docker_dir env_file
	this_dir="$(dirname -- "$(readlink -f -- "${BASH_SOURCE[0]}")")"
	repo_dir="$(readlink -f -- "$this_dir/../..")"
	docker_dir="$repo_dir/docker"
	compose_args=(
		--file="$docker_dir/compose.yml"
		--env-file="$docker_dir/.env" # force error if missing .env
	)
	(
		compose_env
		cd "$docker_dir" || return 1
		set -x
		docker compose "${compose_args[@]}" "$@"
	)
}

# Source `docker/.env` into the current shell environment.
#
# If `docker/.env` does not exist, creates from `docker/util/env.in`.
compose_env() {
	local this_dir repo_dir docker_dir env_file
	this_dir="$(dirname -- "$(readlink -f -- "${BASH_SOURCE[0]}")")"
	repo_dir="$(readlink -f -- "$this_dir/../..")"
	docker_dir="$repo_dir/docker"
	env_file="$docker_dir/.env"
	[ -f "$env_file" ] || _interpolate_env_in "$this_dir/env.in" | _snip >"$env_file"
	# shellcheck source=docker/util/env.in
	. "$env_file"
}

_interpolate_env_in() {
	local path="$1"
	(
		umask 0377 # u=r,go=
		_print_bash_interpreted_file "$path"
	)
}

_print_bash_interpreted_file() {
	local in_file="$1"
	if [ ! -f "$in_file" ]; then
		return 1
	fi
	(
		while IFS='' read -r line || [ -n "$line" ]; do
			#                     || [ -n "$line" ] handles files without
			# trailing newline; otherwise read skips their last line.
			result="$(eval "printf '%s\n' \"$line\"")"
			echo "$result" # print evaluated line
			eval "$result" # preserve side-effects
		done <"$in_file"
	)
}

_snip() {
	sed '/^#[[:space:]]*+snip/,/^#[[:space:]]*-snip/d' |
		cat --squeeze-blank
}
