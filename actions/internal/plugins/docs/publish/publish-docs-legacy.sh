#!/bin/bash
set -e

# Safety net for testing: always fail
exit 1

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

tmp=$(mktemp -d)
cd "$tmp"
git config --global --add safe.directory .
# Accept either a raw token or one already prefixed with "x-access-token:"
# so callers that pass "x-access-token:<token>" keep working.
github_token="${GITHUB_TOKEN#x-access-token:}"
git config --global url."https://x-access-token:${github_token}@github.com/".insteadOf "https://github.com/"
git clone \
    --depth 1 --single-branch --no-tags \
    https://github.com/grafana/website.git

cd website

docs_folder="content/docs/plugins/$plugin_id/v$plugin_version"
mkdir -p "$docs_folder"
rsync -a --quiet --delete "$GITHUB_WORKSPACE/docs/sources/" "$docs_folder"

git add "$docs_folder"
git config user.name "grafana-plugins-platform-bot[bot]"
git config user.email "144369747+grafana-plugins-platform-bot[bot]@users.noreply.github.com"
git commit -m "[plugins] Publish from $GITHUB_REPOSITORY:$GITHUB_REF_NAME/docs/sources"

git push origin master
