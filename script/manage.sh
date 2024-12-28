#!/bin/bash
set -euo pipefail

####################################################################
#  This script forwards arguments to docker/manage.entrypoint.sh,  #
#  running in a Docker container.                                  #
#                                                                  #
#  Try:                                                            #
#    script/manage.sh --help                                       #
#    script/manage.sh -- --help                                    #
####################################################################

script_dir="$(dirname -- "$(readlink -f -- "$0")")"
source_dir="$(readlink --canonicalize "$script_dir/..")"

docker_dir="$source_dir/docker"
compose_yml="$docker_dir/compose.yml"

compose=(docker compose --file "$compose_yml")

"${compose[@]}" build --quiet manage
"${compose[@]}" run --rm manage "$@"
