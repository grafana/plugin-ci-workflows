#!/bin/bash
set -e

# Check if the file was provided as an argument
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <changelog-file>"
  exit 1
fi

# Get the file path
FILE=$1

# Use awk to print the topmost section of the file
awk '
/^## / { 
  if (section) exit; 
  section=1 
} 
section { print }' "$FILE"
