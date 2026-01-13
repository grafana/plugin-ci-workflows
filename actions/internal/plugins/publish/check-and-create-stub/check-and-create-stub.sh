#!/bin/bash
set -e

usage() {
    echo "Usage: $0 --environment <dev|ops|staging> --plugin-id <plugin_id>"
}

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --environment) gcom_env=$2; shift 2;;
        --plugin-id) plugin_id=$2; shift 2;;
        --help)
            usage
            exit 0
            ;;
    esac
done

if [ -z "$plugin_id" ]; then
    echo "Plugin ID not provided."
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

# Allow URL override for testing (GCOM_API_URL takes precedence)
if [ -n "$GCOM_API_URL" ]; then
    gcom_api_url=$GCOM_API_URL
    echo "Using GCOM_API_URL override: $gcom_api_url"
else
    # Can only be 'dev' or 'ops'. The prod stub is created manually.
    case $gcom_env in
        dev)
            gcom_api_url=https://grafana-dev.com/api
            ;;
        ops|staging)
            gcom_api_url=https://grafana-ops.com/api
            ;;
        *)
            echo "Invalid environment: $gcom_env (supported values: 'dev', 'ops')"
            usage
            exit 1
            ;;
    esac
fi

# Build args for curl to GCOM (auth headers)
curl_args=(
    "-H" "Accept: application/json"
    "-H" "User-Agent: github-actions-shared-workflows:/plugins/publish"
)

# Production curl args do not have IAP
prod_curl_args=("${curl_args[@]}")

# Add IAP and API key for dev/ops environments
if [ -z "$GCLOUD_AUTH_TOKEN" ]; then
    echo "GCLOUD_AUTH_TOKEN environment variable not set."
    exit 1
fi
curl_args+=("-H" "Authorization: Bearer $GCLOUD_AUTH_TOKEN")
curl_args+=("-H" "X-Api-Key: $GCOM_PUBLISH_TOKEN")


echo "Checking if a plugin or stub exists for $plugin_id in $gcom_env"
gcom_plugin_code=$(
    curl -sSL \
    "${curl_args[@]}" \
    $gcom_api_url/plugins/$plugin_id \
    | jq -r .code
)
if [ "$gcom_plugin_code" != "NotFound" ]; then
    echo "Plugin or stub already exists for $plugin_id in $gcom_env, nothing to do."
    exit 0
fi

echo "Plugin or stub does not exist in environment $gcom_env, creating it"
echo "Fetching existing signature type from production"
gcom_signature_type=$(
    curl -sSL \
    "${prod_curl_args[@]}" \
    https://grafana.com/api/plugins/$plugin_id \
    | jq -r .signatureType
)
if [ "$gcom_signature_type" == "null" ]; then
    echo "Invalid signature type in prod - make sure the stub exists in grafana.com"
    echo "If this is a brand new plugin, you need to create the stub in grafana.com (prod) first, also when publishing to dev or ops."
    echo "See the documentation for more details:"
    echo "https://enghub.grafana-ops.net/docs/default/component/grafana-plugins-platform/plugins-ci-github-actions/010-plugins-ci-github-actions/#publishing-a-brand-new-plugin-for-the-first-time"
    exit 1
fi
echo "Signature type for $plugin_id is $gcom_signature_type"
echo "Creating plugin stub $plugin_id in $gcom_env with signature type $gcom_signature_type"
out=$(
    curl -sSLf \
    -X POST \
    "${curl_args[@]}" \
    -d stub=true \
    -d slug=$plugin_id \
    -d signatureType=$gcom_signature_type \
    -d orgId=2 \
    $gcom_api_url/plugins
)
echo -e "\nResponse:"
echo $out | jq
