#!/bin/bash
set -e

usage() {
    echo "Usage: $0 --environment <dev|ops|prod> <plugin_zip_urls...>"
}

json_obj() {
    jq -cn "$@" '$ARGS.named'
}

gcs_zip_urls=()
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --environment) gcom_env=$2; shift 2;;
        --help)
            usage
            exit 0
            ;;
        *)
            gcs_zip_urls+=("$1")
            shift
            ;;
    esac
done
echo $gcs_zip_urls

if [ -z "$gcs_zip_urls" ]; then
    echo "Plugin ZIP URLs not provided."
    usage
    exit 1
fi

if [ -z $GCOM_PUBLISH_TOKEN ]; then
    echo "GCOM_PUBLISH_TOKEN environment variable not set."
    exit 1
fi

if [ -z $gcom_env ]; then
    echo "Environment not provided"
    usage
    exit 1
fi

has_iap=false
case $gcom_env in
    dev)
        gcom_api_url=https://grafana-dev.com/api
        has_iap=true
        ;;
    ops)
        gcom_api_url=https://grafana-ops.com/api
        has_iap=true
        ;;
    prod)
        gcom_api_url=https://grafana.com/api
        ;;
    *)
        echo "Invalid environment: $gcom_env (supported values: 'dev', 'ops', 'prod')"
        usage
        exit 1
        ;;
esac

# Build args for curl to GCOM (auth headers)
curl_args=(
    "-H" "Content-Type: application/json"
    "-H" "Accept: application/json"
    "-H" "User-Agent: github-actions-shared-workflows:/plugins/publish"
)
if [ "$has_iap" = true ]; then  
    if [ -z "$GCLOUD_AUTH_TOKEN" ]; then
        echo "GCLOUD_AUTH_TOKEN environment variable not set."
        exit 1
    fi
    curl_args+=("-H" "Authorization: Bearer $GCLOUD_AUTH_TOKEN")
    curl_args+=("-H" "X-Api-Key: $GCOM_PUBLISH_TOKEN")
else
    curl_args+=("-H" "Authorization: Bearer $GCOM_PUBLISH_TOKEN")
fi

# Build JSON payload for publishing
jq_download_args=()
for zip_url in $gcs_zip_urls; do
    platform=any
    os=""
    arch=""

    # Extract os+arch from the file name, if possible
    file=$(basename $zip_url)
    os=$(echo $file | sed -E "s|.+\.(\w+)_\w+\.zip|\1|")
    arch=$(echo $file | sed -E "s|.+\.\w+_(\w+)\.zip|\1|")
    if [ "$file" != "$os" ] && [ "$file" != "$arch" ]; then
        # os-arch zip
        platform="$os-$arch"
    fi
    echo $file

    # Try to get the .md5 for the zip file
    md5_url="$zip_url.md5"
    md5=$(curl --fail -s "$md5_url")
    if [ $? -ne 0 ]; then
        echo "Failed to fetch md5: $md5_url"
        exit 1
    fi
    md5=$(echo $md5 | tr -d '\n')

    # Make sure the md5 is valid length (valid response)
    if [ ${#md5} -ne 32 ]; then
        echo "Invalid md5 ($md5): $md5_url"
        exit 1
    fi

    # Add URL + md5 to JSON payload
    json_artifact=$(json_obj --arg url "$zip_url" --arg md5 "$md5")
    jq_download_args+=("--argjson" "$platform" "$json_artifact")
done

pushd "$GITHUB_WORKSPACE" > /dev/null
sha=$(git rev-parse HEAD)
popd > /dev/null

# Publish the plugin
echo "Publishing to $gcom_api_url..."
json_download=$(json_obj "${jq_download_args[@]}")
json_payload=$(jq -c -n \
    --argjson download "$json_download" \
    --arg url "$GITHUB_SERVER_URL/$GITHUB_REPOSITORY" \
    --arg commit "$sha" \
    '$ARGS.named'
)
echo $json_payload | jq
out=$(
    curl -sSL \
        -X POST \
        "${curl_args[@]}" \
        -d "$json_payload" \
        $gcom_api_url/plugins
)

echo -e "\nResponse:"
set +e
echo $out | jq
if [ $? -ne 0 ]; then
    # Non-JSON output, print raw response
    echo $out
    exit 1
fi

if [[ $(echo "$out" | jq -r '.plugin.id? // empty') != "" ]]; then
    echo -e "\nPlugin published successfully"
else
    echo -e "\nPlugin publish failed"
    exit 1
fi