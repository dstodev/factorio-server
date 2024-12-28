#!/bin/bash
set -euo pipefail

script_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink --canonicalize "$script_dir/..")"

docker_dir="$source_dir/docker"
compose_yml="$docker_dir/compose.yml"

compose=(docker compose --file "$compose_yml")

"${compose[@]}" build --quiet manage
"${compose[@]}" run --rm manage "$@"
