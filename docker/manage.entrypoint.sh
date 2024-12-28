#!/bin/bash
set -euo pipefail

pushd manage >/dev/null

umask 0002

venv_activate() {
	. .venv/bin/activate
}

venv_init() {
	python -m venv .venv &&
		venv_activate &&
		pip install --upgrade pip &&
		pip install --editable .
}

if [ -d .venv ]; then
	venv_activate
else
	venv_init
fi

manage "$@"

popd >/dev/null
