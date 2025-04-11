#!/bin/bash
set -e

original_dir=$(pwd)

tmp=$(mktemp -d)
cp "$1" "$tmp"

pushd "$tmp" > /dev/null
cp -r "$original_dir/$1" .
rm $(basename "$1")

# Ensure we have one folder in the current working directory
if [ $(ls -d */ | wc -l) -ne 1 ]; then
    echo "Expected one folder in the zip file"
    exit 1
fi

# Enter the plugin directory
cd $(ls -d */)

# Ensure we have plugin.json and module.js
for file in plugin.json module.js; do
    if [ ! -f "$file" ]; then
        echo "Expected $file in the zip file"
        exit 1
    fi
done

# Ensure we have a MANIFEST.txt (checking the siguature is too overkill)
if [ ! -f "MANIFEST.txt" ]; then
    echo "Plugin is unsigned, MANIFEST.txt not found"
    exit 1
fi

# If we have an executable, ensure it exists and it's executable
exe=$(jq -r .executable plugin.json)
if [ "$exe" != "null" ]; then
    # Ensure at least one os/arch combo executable exists
    found=false
    # TODO: ensure correct os/arch + exe combo instead (correct name => correct file)
    #   regex: `.+\.(\w+)_(\w+)\.zip`
    for os in linux darwin windows; do
        for arch in amd64 arm64 arm; do
            if [ -f "$exe" ]; then
                if [ ! -x "$exe" ]; then
                    echo "Executable $exe is not flagged as executable"
                    exit 1
                fi
                found=true
                break 2
            fi
        done
    done

    if [ "$found" = false ]; then
        echo "No executable found in the zip file"
        exit 1
    fi
fi

# TODO: support nested apps

popd > /dev/null
