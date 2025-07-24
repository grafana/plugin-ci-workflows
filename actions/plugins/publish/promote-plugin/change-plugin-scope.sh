#!/bin/bash
set -e
if [ "$RUNNER_DEBUG" == "1" ]; then
    set -x
fi

usage() {
    echo "Usage: $0 --environment <dev|ops|prod> [--dry-run] <plugin-id> <plugin-version>"
}

json_obj() {
    jq -cn "$@" '$ARGS.named'
}

dry_run=false
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --environment) gcom_env=$2; shift 2;;
        --dry-run) dry_run=true; shift;;
        --help)
            usage
            exit 0
            ;;
        *)
            plugin_id=$1
            plugin_version=$2
            shift
            ;;
    esac
done

if [ -z $GCOM_PUBLISH_TOKEN ]; then
    echo "GCOM_PUBLISH_TOKEN environment variable not set."
    exit 1
fi

if [ -z $gcom_env ]; then
    echo "Environment not provided"
    usage
    exit 1
fi

if [ -z $plugin_id ]; then
    echo "Plugin ID not provided"
    usage
    exit 1
fi

if [ -z $plugin_version ]; then
    echo "Version not provided"
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

# Create a json payload that has a property "scopes": ["universal"]
json_payload=$(json_obj --argjson scopes '["universal"]')

echo $json_payload | jq
if [ "$dry_run" = true ]; then
    echo "Dry run enabled, skipping publish"
    exit 0
fi
out=$(
    curl -sSL \
        -X POST \
        "${curl_args[@]}" \
        -d "$json_payload" \
        $gcom_api_url/plugins/$plugin_id/versions/$plugin_version
)

echo -e "\nResponse:"
set +e
echo $out | jq
if [ $? -ne 0 ]; then
    # Non-JSON output, print raw response
    echo $out
    exit 1
fi

# Determine if publish succeeded
if [[ $(echo "$out" | jq -r '.scope[]? | select(. == "universal")') == "universal" ]]; then
    echo -e "\nPlugin scopes successfully changed"
else
    echo -e "\nPlugin publish failed"
    exit 1
fi
