#!/bin/bash
set -euo pipefail

validate_plugin_zip() {
    local zip_url="$1"
    local plugin_id="$2"
    local temp_dir=$(mktemp -d)
    local zip_file="$temp_dir/plugin.zip"

    echo "Processing: $zip_url"

    # Download the ZIP file
    if ! curl -L -o "$zip_file" "$zip_url"; then
        echo "ERROR: Failed to download $zip_url"
        return 1
    fi

    # Extract ZIP file
    if ! unzip -q "$zip_file" -d "$temp_dir"; then
        echo "ERROR: Failed to extract $zip_url"
        return 1
    fi

    # Check if plugin_id folder exists
    if [[ ! -d "$temp_dir/$plugin_id" ]]; then
        echo "ERROR: Folder '$plugin_id' not found in $zip_url"
        return 1
    fi

    # Check if MANIFEST.txt exists and is non-empty
    if [[ ! -f "$temp_dir/$plugin_id/MANIFEST.txt" ]] || [[ ! -s "$temp_dir/$plugin_id/MANIFEST.txt" ]]; then
        echo "ERROR: MANIFEST.txt missing or empty in $zip_url"
        return 1
    fi

    # Check if plugin.json exists and is non-empty
    if [[ ! -f "$temp_dir/$plugin_id/plugin.json" ]] || [[ ! -s "$temp_dir/$plugin_id/plugin.json" ]]; then
        echo "ERROR: plugin.json missing or empty in $zip_url"
        return 1
    fi

    echo "SUCCESS: $zip_url passed all validation checks"
    return 0
}

process_plugin_zips() {
    local urls_array="$1"
    local plugin_id="$2"
    local failed_count=0
    local zip_urls

    # Parse JSON array; -e also rejects an empty array
    if ! zip_urls=$(echo "$urls_array" | jq -er '.[]'); then
        echo "ERROR: No ZIP URLs to validate in: $urls_array"
        return 1
    fi

    # Read from a here-string, not a pipe, so failed_count isn't updated in a subshell
    while read -r zip_url; do
        if ! validate_plugin_zip "$zip_url" "$plugin_id"; then
            failed_count=$((failed_count + 1))
        fi
        echo "---"
    done <<< "$zip_urls"

    if [[ $failed_count -gt 0 ]]; then
        echo "Completed with $failed_count failures"
        return 1
    else
        echo "All ZIP files validated successfully"
        return 0
    fi
}

# Main script execution
if [[ $# -ne 2 ]]; then
    echo "Usage: $0 <json_array_of_urls_or_single_url> <plugin_id>"
    exit 1
fi
urls_array="$1"
plugin_id="$2"

# If a string was provided in urls_array, convert it to a JSON array with jq
if [[ "$urls_array" != \[* ]]; then
    urls_array=$(echo "$urls_array" | jq -Rc '[.]')
fi

if ! process_plugin_zips "$urls_array" "$plugin_id"; then
    exit 1
fi
exit 0