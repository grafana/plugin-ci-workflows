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

if [ -f .nvmrc ]; then
    nvm use
    echo "Using Node version: $(node -v)"
else
    echo "No .nvmrc file found, using default Node version from simple-frontend testdata"
    nvm use $(cat ../simple-frontend/.nvmrc)
fi

echo "Using Node version: $(node -v)"
echo "Using Package Manager: $(pm)"

echo "[$1] (frontend) Installing"
$(pm) install

echo "[$1] (frontend) Building the plugin"
$(pm) run build

if [ -f "Magefile.go" ]; then
    echo "[$1] (backend) Building the plugin"
    go mod download -x
    mage -v buildAll
else
    echo "[$1] No backend to build"
fi

echo "[$1] Copying dist folder to mockdata"
rm -rf "../act/mockdata/dist/$1"
mkdir -p "../act/mockdata/dist/$1"
cp -r dist/* "../act/mockdata/dist/$1/"
