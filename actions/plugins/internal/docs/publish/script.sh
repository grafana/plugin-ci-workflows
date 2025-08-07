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

plugin_id="$1"
plugin_version="$2"

tmp=$(mktemp -d)
cd "$tmp"
git config --global --add safe.directory .
git clone \
    --depth 1 --single-branch --no-tags \
    https://$GITHUB_TOKEN@github.com/grafana/website.git

cd website

docs_folder="content/docs/plugins/$plugin_id/v$plugin_version"
mkdir -p "$docs_folder"
cp -a "$GITHUB_WORKSPACE/docs/sources/." "$docs_folder/"

git add "$docs_folder"
git config user.name "144369747+grafana-plugins-platform-bot[bot]@users.noreply.github.com"
git config user.email "grafana-plugins-platform-bot[bot]"
git commit -m "[plugins] Publish from $GITHUB_REPOSITORY:$GITHUB_REF_NAME/docs/sources"

git push origin master
