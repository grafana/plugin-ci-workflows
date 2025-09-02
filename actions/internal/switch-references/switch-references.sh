#!/bin/bash

set -euo pipefail

# Check if replacement parameter is provided
if [[ $# -eq 0 ]]; then
    echo "Usage: $0 <replacement_string> [-v]"
    echo ""
    echo "Replaces all 'uses' references for grafana/plugin-ci-workflows"
    echo "from @<current_version> to @<replacement_string>"
    echo ""
    echo "Arguments"
    echo "  -v: verbose"
    echo ""
    echo "Example: $0 main"
    echo "Example: $0 v2.0.0"
    exit 1
fi

REPO_NAME=grafana/plugin-ci-workflows
VERBOSE=false
if [ $# -gt 1 ] && [ "$2" == "-v" ]; then
    VERBOSE=true
fi

echov() {
    if [ "$VERBOSE" == true ]; then
        echo "$@"
    fi
}

# Get the replacement string from the first argument
REPLACEMENT="$1"

# Function to process YAML files in a directory
process_directory() {
    local dir="$1"

    # Check if directory exists
    if [[ ! -d "$dir" ]]; then
        echov "Directory $dir does not exist, skipping..."
        return
    fi

    # Find all .yml and .yaml files in the directory and subdirectories
    find "$dir" -type f \( -name "*.yml" -o -name "*.yaml" \) -print0 | while IFS= read -r -d '' file; do
        echov "Processing: $file"

        # Check if file contains the pattern before modifying
        if grep -q "uses: $REPO_NAME.*@" "$file"; then
            # Use sed to replace the pattern in-place while preserving comments
            # Pattern explanation:
            # - \(uses: $REPO_NAME[^@]*\) - Capture everything up to @
            # - @[^ ]* - Match @ and everything up to the first space (or end of line)
            # - \(.*\) - Capture everything after the version (including comments)
            # Replace with: captured prefix + @REPLACEMENT + captured suffix
            sed -i "s|\(uses: $REPO_NAME[^@]*\)@[^ ]*\(.*\)|\1@$REPLACEMENT\2|g" "$file"
            echo "  ✓ Updated $file"
        else
            echov "  - No matching pattern found in $file"
        fi
    done
}

# Main script
echo "Starting YAML file processing..."
echo "Replacement string: @$REPLACEMENT"
echo "================================"

# Process the specified directories
process_directory ".github/workflows"
process_directory "actions/plugins"

echo "================================"
echo "Processing complete!"
