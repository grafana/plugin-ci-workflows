#!/bin/bash
set -e
if [[ "$RUNNER_DEBUG" == "1" ]]; then
    set -x
fi

usage() {
    echo "Usage: $0 --environment <dev|ops|prod> --scopes <universal|grafana_cloud|a,b,c|...> [--dry-run] <plugin-id> <plugin-version>"
}

json_obj() {
    jq -cn "$@" '$ARGS.named'
}

dry_run=false
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --environment) gcom_env=$2; shift 2;;
        --scopes) scopes=$(echo $2 | jq -Rc 'split(",")'); shift 2;;
        --dry-run) dry_run=true; shift;;
        --help)
            usage
            exit 0
            ;;
        *)
            # Break on first non-flag argument to handle positional args
            break
            ;;
    esac
done

# Handle remaining positional arguments
plugin_id=$1
plugin_version=$2


if [ -z $GCOM_TOKEN ]; then
    echo "GCOM_TOKEN environment variable not set."
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

if [ -z $scopes ]; then
    echo "Scopes not provided"
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
    curl_args+=("-H" "X-Api-Key: $GCOM_TOKEN")
else
    curl_args+=("-H" "Authorization: Bearer $GCOM_TOKEN")
fi

json_payload=$(json_obj --argjson scopes "$scopes")

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

check_response_error() {
    local response="$1"
    local expected_scopes="$2"

    local response_scopes_count
    response_scopes_count=$(echo "$response" | jq -r '(.scopes // []) | length')
    if [[ "$response_scopes_count" == "0" ]]; then
        echo -e "\nPlugin scopes changed failed"
        return 1
    fi

    local response_scopes
    response_scopes=$(echo "$response" | jq -r '.scopes // []')

    IFS=',' read -ra expected_scopes_array <<< "$expected_scopes"
    for expected_scope in "${expected_scopes_array[@]}"; do
        # Remove leading and trailing whitespace from expected scope
        expected_scope=$(echo "$expected_scope" | xargs)

        local scope_found
        scope_found=$(echo "$response_scopes" | jq --arg scope "$expected_scope" -r 'index($scope) != null')
        if [[ "$scope_found" != "true" ]]; then
            echo -e "\nPlugin scopes changed failed - missing scope: $expected_scope"
            return 1
        fi
    done

    return 0
}

check_response_error "$out" "$scopes"
exit_code=$?
if [[ $exit_code -eq 0 ]]; then
    echo -e "\nPlugin scopes successfully changed"
else
    exit $exit_code
fi
