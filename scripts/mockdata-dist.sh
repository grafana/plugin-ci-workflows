#!/bin/bash
set -euo pipefail

export NVM_DIR=$HOME/.nvm;
source $NVM_DIR/nvm.sh;

pm() {
    if [ -f yarn.lock ]; then
        echo "yarn"
    elif [ -f pnpm-lock.yaml ]; then
        echo "pnpm"
    elif [ -f package-lock.json ]; then
        echo "npm"
    else
        echo "No recognized package manager found in this project."
        exit 1
    fi
}

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <test-plugin-folder-name>"
    exit 1
fi

echo "[$1] Preparing mockdata (dist)"
cd "$(dirname "$0")/.."
cd tests/$1

nvm use
echo "Using Node version: $(node -v)"
echo "Using Package Manager: $(pm)"

echo "[$1] (frontend) Installing"
$(pm) install

echo "[$1] (frontend) Lint + typecheck"
$(pm) run lint
$(pm) run typecheck

echo "[$1] (frontend) Building the plugin"
$(pm) run build

if [ -f "Magefile.go" ]; then
    echo "[$1] (backend) Building the plugin"
    go mod download -x
    mage -v buildAll
else
    echo "[$1] No backend to build"
fi

if [ -n "$(jq -r '.docsPath // empty' src/plugin.json)" ]; then
    docs_cli_version=$(sed -n 's/^ *DEFAULT_PLUGIN_DOCS_CLI_VERSION: *"\(.*\)"/\1/p' ../../.github/workflows/ci.yml)
    echo "[$1] Building catalog docs with @grafana/plugin-docs-cli@${docs_cli_version}"
    npx --yes "@grafana/plugin-docs-cli@${docs_cli_version}" validate --strict
    npx --yes "@grafana/plugin-docs-cli@${docs_cli_version}" build
else
    echo "[$1] No catalog docs to build"
fi

echo "[$1] Copying dist folder to mockdata"
rm -rf "../act/mockdata/dist/$1"
mkdir -p "../act/mockdata/dist/$1"
cp -r dist/* "../act/mockdata/dist/$1/"
