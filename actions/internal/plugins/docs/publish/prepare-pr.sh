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
# Mask the stripped token and expose it so the create-pull-request step can authenticate.
echo "::add-mask::${github_token}"
echo "token=${github_token}" >> "$GITHUB_OUTPUT"

# create-pull-request requires the repository to live under GITHUB_WORKSPACE,
# so clone it into a subdirectory there rather than a temp dir.
clone_dir="_website-publish"
abs_clone_dir="${GITHUB_WORKSPACE}/${clone_dir}"
rm -rf "$abs_clone_dir"

git config --global --add safe.directory "$abs_clone_dir"
git config --global url."https://x-access-token:${github_token}@github.com/".insteadOf "https://github.com/"
git clone \
    --depth 1 --single-branch --no-tags \
    https://github.com/grafana/website.git "$abs_clone_dir"

docs_folder="${abs_clone_dir}/content/docs/plugins/$plugin_id/v$plugin_version"
mkdir -p "$docs_folder"
rsync -a --quiet --delete "$GITHUB_WORKSPACE/docs/sources/" "$docs_folder"

# Expose the clone directory, relative to GITHUB_WORKSPACE, for the
# create-pull-request step's `path` input. create-pull-request stages and
# commits the rsync'd changes itself, so no `git add`/commit here.
echo "dir=${clone_dir}" >> "$GITHUB_OUTPUT"
