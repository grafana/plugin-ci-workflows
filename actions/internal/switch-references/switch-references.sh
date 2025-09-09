#!/bin/bash

set -euo pipefail

echov() {
    if [ "$VERBOSE" == true ]; then
        echo "$@"
    fi
}

show_help() {
    echo "Usage: $0 [-v|--verbose] <repository_name> <replacement_string> <paths...>"
    echo ""
    echo "Replaces all 'uses' references for the specified repository"
    echo "from @<current_version> to @<replacement_string> into each of the"
    echo "specified paths."
    echo ""
    echo "If a path is a directory, all .yml/.yaml files will be processed"
    echo "recursively. If a path is a file, only that file will be processed."
    echo ""
    echo "Arguments"
    echo "  -v: verbose"
    echo ""
    echo "Examples:"
    echo "  $0 grafana/plugin-ci-workflows main .github/workflows          # Process directory recursively"
    echo "  $0 grafana/plugin-ci-workflows v2.0.0 .github/workflows/ci.yml # Process single file"
    echo "  $0 grafana/plugin-ci-workflows main examples actions/plugins   # Process multiple directories"
}


# Function to process a single file
process_file() {
    local file="$1"
    
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
        echo "  - No matching pattern found in $file"
    fi
}

# Function to process YAML files
process() {
    local path="$1"

    # Check if path is a directory (treat as recursive search)
    if [[ -d "$path" ]]; then
        echov "Processing directory: $path"
        
        # Find all .yml and .yaml files in the directory and subdirectories
        find "$path" -type f \( -name "*.yml" -o -name "*.yaml" \) -print0 | while IFS= read -r -d '' file; do
            process_file "$file"
        done
    elif [[ -f "$path" ]]; then
        # Check if it's a YAML file
        if [[ "$path" == *.yml || "$path" == *.yaml ]]; then
            process_file "$path"
        else
            echov "File $path is not a YAML file, skipping..."
        fi
    else
        echo "Path $path does not exist, skipping..."
    fi
}

# Arguments parsing
VERBOSE=false
paths=()
for arg in "$@"; do
    case $arg in
        -h|--help) show_help; exit 0 ;;
        -v|--verbose) VERBOSE=true;;
        *)
            if [ -z "${REPO_NAME:-}" ]; then
                REPO_NAME="$arg"
            elif [ -z "${REPLACEMENT:-}" ]; then
                REPLACEMENT="$arg"
            else
                paths+=("$arg")
            fi
    esac
done

if [ -z "${REPO_NAME:-}" ] || [ -z "${REPLACEMENT:-}" ] || [ ${#paths[@]} -eq 0 ]; then
    # Not enough arguments
    echo "Error: Missing required arguments."
    show_help
    exit 1
fi

# Main script
if [[ ! "$REPO_NAME" =~ ^[a-zA-Z0-9._/-]+$ ]]; then
    echo "Error: Invalid repository name" >&2
    exit 1
fi
# Escape special characters in REPO_NAME for use in sed
REPO_NAME=$(printf '%s\n' "$REPO_NAME" | sed 's/[[\.*^$()+?{|]/\\&/g')

echo "Starting YAML file processing..."
echo "Searching for repository: $REPO_NAME"
echo "Replacing its references with: @$REPLACEMENT"
echo "================================"

# Process each path provided as argument
for path in "${paths[@]}"; do
    process "$path"
done

echo "================================"
echo "Processing complete!"
