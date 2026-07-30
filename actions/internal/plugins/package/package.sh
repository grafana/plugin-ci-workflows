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

# Discover every backend executable family declared by a packaged plugin.json.
# Executable paths are relative to the plugin.json that declares them.
backend_executables=()
while IFS= read -r -d '' plugin_json; do
    executable=$(jq -r '.executable // empty' "$plugin_json")
    if [ -z "$executable" ]; then
        continue
    fi

    plugin_dir=$(dirname "$plugin_json")
    if [ "$plugin_dir" = "." ]; then
        backend_executables+=("$executable")
    else
        backend_executables+=("${plugin_dir#./}/$executable")
    fi
done < <(find . -type f -name plugin.json -print0)

# Collect all executable files and the unique os/arch variants they provide.
backend_files=()
os_arches=()
for executable in "${backend_executables[@]}"; do
    executable_dir=$(dirname "$executable")
    executable_basename=$(basename "$executable")

    while IFS= read -r -d '' file; do
        file=${file#./}
        filename=$(basename "$file")
        os_arch=${filename#"${executable_basename}_"}
        os_arch=${os_arch%.exe}
        if [[ ! "$os_arch" =~ ^[a-zA-Z0-9_]+$ ]]; then
            continue
        fi

        backend_files+=("$file")
        os_arches+=("$os_arch")
    done < <(find "$executable_dir" -maxdepth 1 -type f -name "${executable_basename}_*" -print0)
done

if [ "${#backend_files[@]}" -eq 0 ]; then
    echo "No executable found in plugin.json"
    exit 0
fi

unique_os_arches=()
while IFS= read -r os_arch; do
    if [ -n "$os_arch" ]; then
        unique_os_arches+=("$os_arch")
    fi
done < <(printf '%s\n' "${os_arches[@]}" | sort -u)

selected_executable_path() {
    local executable=$1
    local os_arch=$2
    local candidate="$dist/${executable}_${os_arch}"

    if [ -f "$candidate" ]; then
        printf '%s\n' "$candidate"
    elif [ -f "${candidate}.exe" ]; then
        printf '%s\n' "${candidate}.exe"
    else
        return 1
    fi
}

# A tailored archive is only valid when every backend family supports its
# platform. Fail before packaging instead of publishing an incomplete plugin.
for os_arch in "${unique_os_arches[@]}"; do
    for executable in "${backend_executables[@]}"; do
        if ! selected_executable_path "$executable" "$os_arch" > /dev/null; then
            echo "Executable '$executable' does not provide $os_arch, aborting."
            exit 1
        fi
    done
done

# Create one zip per unique os+arch combination. Each zip copies all non-backend
# files, then adds the selected executable from every backend family.
for os_arch in "${unique_os_arches[@]}"; do
    tmp=$(mktemp -d)
    pushd "$tmp" > /dev/null
    mkdir -p "$plugin_id"

    pushd "$dist" > /dev/null
    while IFS= read -r -d '' file; do
        relative_file=${file#./}
        is_backend_file=false
        for backend_file in "${backend_files[@]}"; do
            if [ "$relative_file" = "$backend_file" ]; then
                is_backend_file=true
                break
            fi
        done
        if [ "$is_backend_file" = true ]; then
            continue
        fi

        dir=$(dirname "$file")
        mkdir -p "$tmp/$plugin_id/$dir"
        cp -p "$file" "$tmp/$plugin_id/$dir/"
    done < <(find . -type f -print0)
    popd > /dev/null

    for executable in "${backend_executables[@]}"; do
        selected_executable=$(selected_executable_path "$executable" "$os_arch")
        executable_dest="$tmp/$plugin_id/$(dirname "$executable")"
        mkdir -p "$executable_dest"
        cp -p "$selected_executable" "$executable_dest/"
    done

    os_arch_zip_fn="$plugin_id-$plugin_version.$os_arch.zip"
    echo "Creating package: $os_arch_zip_fn"
    package "$out/$os_arch_zip_fn" "$plugin_id" "$SIGNATURE_TYPE"

    popd > /dev/null
    rm -rf "$tmp"
done
