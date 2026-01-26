#!/bin/bash
set -euo pipefail

hot_dir="$1"

cd "$hot_dir"

# https://support.hytale.com/hc/en-us/articles/45326769420827-Hytale-Server-Manual
wget --output-document="$hot_dir/pkg.zip" 'https://downloader.hytale.com/hytale-downloader.zip'
unzip pkg.zip
rm pkg.zip
