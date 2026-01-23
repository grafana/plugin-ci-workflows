#!/bin/bash
set -e

# package creates a zip file and its corresponding sha256 file
# $1 is the destination zip file name
# $2 is the source directory
# $3 is the signature type (optional, defaults to "grafana")
package() {
    local signature_type=${3:-grafana}
    
    # Sign the plugin
    if [ ! -z $GRAFANA_ACCESS_POLICY_TOKEN ]; then
        npx -y @grafana/sign-plugin@latest --signatureType=$signature_type --distDir $2
    else
        echo "WARNING: Plugin won't be signed, GRAFANA_ACCESS_POLICY_TOKEN not set"
    fi

    zip -r $1 $2
    sha1sum $1 | cut -f1 -d' ' | tr -d '\n' > $1.sha1
    md5sum $1 | cut -f1 -d' ' | tr -d '\n' > $1.md5
}

universal=false
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -u|--universal) universal=true; shift ;;
        *)
            if [ -z "$dist" ]; then
                dist=$(realpath $1)
            elif [ -z "$out" ]; then
                out=$(realpath $1)
            else
                echo "Too many parameters"
                exit 1
            fi
            shift
            ;;
    esac
done

if [ -z "$dist" ] || [ -z "$out" ]; then
    echo "Usage: $0 [-u|--universal] <input_folder_name> <output_folder_name>"
    exit 1
fi

if [ ! -d "$dist" ]; then
    echo "Input folder '$dist' not found, aborting."
    exit 1
fi

if [ ! -f "$dist/plugin.json" ]; then
    echo "plugin.json not found in input folder '$dist', aborting."
    exit 1
fi

mkdir -p $out

cd $dist
plugin_id=$(jq -r .id plugin.json)
plugin_version=$(jq -r .info.version plugin.json)
if [ -z "$plugin_id" ] || [ -z "$plugin_version" ]; then
    echo "plugin.json is missing id or version, aborting."
    exit 1
fi

# Create universal zip (all os+arch combos)
if [ "$universal" = true ]; then
    universal_zip_fn=$plugin_id-$plugin_version.zip
    echo "Creating universal package: $universal_zip_fn"

    tmp=$(mktemp -d)
    mkdir -p "$tmp/$plugin_id"

    cp -r . "$tmp/$plugin_id"
    cd "$tmp"
    package "$out/$universal_zip_fn" "$plugin_id" "$SIGNATURE_TYPE"
    exit 0
fi

# Identify apps with nested datasource
ptype=$(jq -r .type plugin.json)
exe=$(jq -r .executable plugin.json)
backend_folder="."
if [ "$ptype" == "app" ] && [ -d "datasource" ]; then
    echo "Found nested datasource"
    cd datasource
    nested_exe=$(jq -r .executable plugin.json)
    if [ "$nested_exe" != "null" ]; then
        backend_folder="datasource"
        exe="$backend_folder/$nested_exe"
    fi
    cd ..
fi

if [ "$exe" == "null" ]; then
    echo "No executable found in plugin.json"
    exit 0
fi

# Create os+arch zips
exe_basename=$(basename $exe)
for file in $(find "$backend_folder" -type f -name "${exe_basename}_*"); do
    # Extract os+arch from the file name
    os_arch=$(echo $(basename $file) | sed -E "s|${exe_basename}_([a-zA-Z0-9_]+)(.exe)?|\1|")

    # Temporary folder for the zip file
    tmp=$(mktemp -d)
    pushd $tmp > /dev/null
    mkdir -p "$plugin_id"

    # Copy all files but the executables, preserving permissions and mod times (similar to rsync)
    pushd "$dist" > /dev/null
    # -name "${exe_basename}*" -prune: Ignore (prune) all executables
    # -o -type f -print0: OR, print file name (NUL-terminated) for use with while read
    # Copy with cp, preserving permissions and create any required parent directories to the dest folder
    # Note: Using a while loop instead of xargs for macOS compatibility (BSD cp lacks --parents and -t)
    find . -name "${exe_basename}*" -prune -o -type f -print0 | while IFS= read -r -d '' file; do
        dir=$(dirname "$file")
        mkdir -p "$tmp/$plugin_id/$dir"
        cp -p "$file" "$tmp/$plugin_id/$dir/"
    done
    popd > /dev/null

    # Copy only the current executable
    cp "$dist/$file" "$tmp/$plugin_id/$backend_folder"
    os_arch_zip_fn="$plugin_id-$plugin_version.$os_arch.zip"
    echo "Creating package: $os_arch_zip_fn"

    # Create the zip+sha256 files
    package "$out/$os_arch_zip_fn" "$plugin_id" "$SIGNATURE_TYPE"

    # Cleanup temporary folder
    popd > /dev/null
    rm -rf $tmp
done

