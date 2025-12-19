#!/bin/bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <test-plugin-folder-name>"
    exit 1
fi

echo "[$1] Preparing mockdata (dist-artifacts)"
cd "$(dirname "$0")/.."


mkdir -p "tests/act/mockdata/dist-artifacts-unsigned/$1"

echo "[$1] Packaging os/arch ZIPs"
# Will exit with 0 if the plugin has no backend
# (in that case, there's no need for os/arch ZIPs, just universal)
./actions/internal/plugins/package/package.sh "tests/act/mockdata/dist/$1" "tests/act/mockdata/dist-artifacts-unsigned/$1"

echo "[$1] Packaging universal ZIPs"
./actions/internal/plugins/package/package.sh -u "tests/act/mockdata/dist/$1" "tests/act/mockdata/dist-artifacts-unsigned/$1"