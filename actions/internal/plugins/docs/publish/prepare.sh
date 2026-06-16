#!/bin/bash
set -e

if [ $# -ne 2 ]; then
    echo "Usage: $0 <plugin_id> <plugin_version>"
    exit 1
fi

if [ -z "$GITHUB_TOKEN" ]; then
    echo "GITHUB_TOKEN is not set"
    exit 1
fi

echo "Installing pre-requisites"
apk add rsync

plugin_id="$1"
plugin_version="$2"

# Accept either a raw token or one already prefixed with "x-access-token:"
# so callers that pass "x-access-token:<token>" keep working.
github_token="${GITHUB_TOKEN#x-access-token:}"
# Mask the stripped token and expose it so the API commit step can authenticate.
echo "::add-mask::${github_token}"
echo "token=${github_token}" >> "$GITHUB_OUTPUT"

tmp=$(mktemp -d)
cd "$tmp"
git config --global --add safe.directory .
git config --global url."https://x-access-token:${github_token}@github.com/".insteadOf "https://github.com/"
git clone \
    --depth 1 --single-branch --no-tags \
    https://github.com/grafana/website.git

cd website

docs_folder="content/docs/plugins/$plugin_id/v$plugin_version"
mkdir -p "$docs_folder"
rsync -a --quiet --delete "$GITHUB_WORKSPACE/docs/sources/" "$docs_folder"

# Stage all changes so the API commit step can read them via `git status`.
git add -A

# Expose the website clone directory to the API commit step.
echo "dir=${tmp}/website" >> "$GITHUB_OUTPUT"
