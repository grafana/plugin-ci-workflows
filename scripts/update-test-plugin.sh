#!/usr/bin/env bash
#
# Updates dependencies and tooling for a single test plugin under tests/.
# Designed to be invoked one step at a time by the /update-test-plugins slash
# command (or run end-to-end with --all for a non-interactive sweep).
#
# Usage:
#   scripts/update-test-plugin.sh <plugin> --step <step>
#   scripts/update-test-plugin.sh <plugin> --all
#
# Available steps (run in this order by --all):
#   node-version     Update .nvmrc to latest LTS major; install via nvm
#   pkg-manager      Update packageManager field; activate via corepack
#   cp-update        Run `npx @grafana/create-plugin@latest update --force`
#   pin-exact        Rewrite all non-exact versions in package.json to exact
#   list-grafana     List @grafana/* packages in package.json (for Claude)
#   go-version       Bump `go` directive in go.mod to latest stable
#   go-sdk           Bump grafana-plugin-sdk-go to latest
#   osv-scan         Run osv-scanner and print JSON report
#   go-tidy          Run `go mod tidy`
#   verify           Install, typecheck, lint, build (+ mage if backend)
#
# Notes:
#   - Yarn is intentionally pinned to the latest 1.x release. The v4 migration
#     is pending; revisit this when ready.
#   - All package.json version values are kept as exact semver to reduce
#     supply-chain risk. The pin-exact step enforces this invariant.

set -euo pipefail

# --- bootstrap ----------------------------------------------------------------

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# Source nvm so node/npm/pnpm/yarn invocations pick up the right Node version.
# shellcheck disable=SC1091
export NVM_DIR="${NVM_DIR:-$HOME/.nvm}"
if [ -s "$NVM_DIR/nvm.sh" ]; then
    . "$NVM_DIR/nvm.sh"
else
    echo "ERROR: nvm not found at $NVM_DIR/nvm.sh" >&2
    exit 1
fi

# --- arg parsing --------------------------------------------------------------

usage() {
    sed -n '2,29p' "$0" >&2
    exit "${1:-1}"
}

if [ "$#" -lt 1 ]; then
    usage 1
fi

PLUGIN="$1"
shift

STEP=""
RUN_ALL=0

while [ "$#" -gt 0 ]; do
    case "$1" in
        --step)
            [ "$#" -ge 2 ] || { echo "ERROR: --step requires a value" >&2; exit 1; }
            STEP="$2"
            shift 2
            ;;
        --all)
            RUN_ALL=1
            shift
            ;;
        -h|--help)
            usage 0
            ;;
        *)
            echo "ERROR: unknown argument: $1" >&2
            exit 1
            ;;
    esac
done

if [ "$RUN_ALL" -eq 0 ] && [ -z "$STEP" ]; then
    echo "ERROR: must specify either --step <name> or --all" >&2
    exit 1
fi

# --- validate plugin ----------------------------------------------------------

# Use find-tests.sh to enumerate valid plugins so we never hardcode names and
# automatically pick up new test plugins added in the future. Dotfile dirs
# (e.g. .claude/) are not plugins and are excluded here defensively.
list_plugins() {
    ./scripts/find-tests.sh | sed 's|^\./||' | grep -vE '^\.' || true
}

VALID=0
while IFS= read -r name; do
    [ -z "$name" ] && continue
    if [ "$name" = "$PLUGIN" ]; then
        VALID=1
        break
    fi
done < <(list_plugins)

if [ "$VALID" -ne 1 ]; then
    echo "ERROR: '$PLUGIN' is not a valid test plugin." >&2
    echo "Valid plugins:" >&2
    list_plugins | sed 's/^/  /' >&2
    exit 1
fi

PLUGIN_DIR="$REPO_ROOT/tests/$PLUGIN"

# --- helpers ------------------------------------------------------------------

log() {
    printf '[%s] %s\n' "$PLUGIN" "$*" >&2
}

require_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "ERROR: required command '$1' not found in PATH" >&2
        exit 1
    fi
}

# Detect the package manager for the current plugin via lockfile presence.
detect_pm() {
    if [ -f "$PLUGIN_DIR/yarn.lock" ]; then
        echo "yarn"
    elif [ -f "$PLUGIN_DIR/pnpm-lock.yaml" ]; then
        echo "pnpm"
    elif [ -f "$PLUGIN_DIR/package-lock.json" ]; then
        echo "npm"
    else
        # Fall back to packageManager field if no lockfile yet.
        local pm_field
        pm_field="$(jq -r '.packageManager // ""' "$PLUGIN_DIR/package.json" 2>/dev/null || true)"
        case "$pm_field" in
            yarn@*) echo "yarn" ;;
            pnpm@*) echo "pnpm" ;;
            npm@*)  echo "npm"  ;;
            *) echo "npm" ;;
        esac
    fi
}

# Fetch the latest stable Go version (without the leading "go") from go.dev.
latest_go_version() {
    require_cmd curl
    require_cmd jq
    curl -fsSL 'https://go.dev/dl/?mode=json' \
        | jq -r '[.[] | select(.stable==true)] | .[0].version' \
        | sed 's/^go//'
}

# Fetch the latest LTS major Node version (e.g. "22" or "24").
latest_lts_major() {
    # nvm ls-remote --lts prints lines like "v22.11.0   (LTS: Jod)   *"
    # We grab the last entry which is the newest.
    local v
    v="$(nvm ls-remote --lts --no-colors 2>/dev/null | grep -E '^\s*v[0-9]+' | tail -1 | awk '{print $1}')"
    if [ -z "$v" ]; then
        echo "ERROR: failed to determine latest LTS Node version" >&2
        exit 1
    fi
    # strip leading "v" and keep only the major
    v="${v#v}"
    echo "${v%%.*}"
}

# Latest published version of an npm package on the default dist-tag.
npm_latest() {
    local pkg="$1"
    npm view "$pkg" version 2>/dev/null
}

# Highest version of a package matching a constraint (e.g. "1" -> latest 1.x).
# Uses npm view + plain text parsing. Returns the last (highest) match.
npm_latest_matching() {
    local pkg="$1"
    local range="$2"
    npm view "${pkg}@${range}" version 2>/dev/null | tail -1 | awk '{print $2}' | tr -d "'\""
}

# Latest package manager version, respecting our policy (yarn stays on 1.x).
latest_pm_version() {
    local pm="$1"
    case "$pm" in
        npm|pnpm)
            npm_latest "$pm"
            ;;
        yarn)
            # Yarn 1.x maintenance line. TODO: revisit when migrating to v4.
            npm_latest_matching yarn '1'
            ;;
        *)
            echo "ERROR: unsupported package manager: $pm" >&2
            exit 1
            ;;
    esac
}

# Run a command inside the plugin directory with the plugin's Node active.
in_plugin() {
    (
        cd "$PLUGIN_DIR"
        # Activate Node per .nvmrc; install if missing.
        if [ -f .nvmrc ]; then
            nvm install >/dev/null 2>&1 || true
            nvm use >/dev/null 2>&1
        fi
        "$@"
    )
}

# Ensure .npmrc contains save-exact=true so any future <pm> add stays exact.
ensure_save_exact_npmrc() {
    local f="$PLUGIN_DIR/.npmrc"
    if [ ! -f "$f" ]; then
        printf 'save-exact=true\n' > "$f"
        return
    fi
    if ! grep -qE '^save-exact[[:space:]]*=' "$f"; then
        printf 'save-exact=true\n' >> "$f"
    fi
}

# --- step implementations -----------------------------------------------------

step_node_version() {
    log "Updating .nvmrc to latest LTS major"
    local major
    major="$(latest_lts_major)"
    echo "$major" > "$PLUGIN_DIR/.nvmrc"
    log "  .nvmrc -> $major"
    # Install + activate so subsequent steps use it.
    (cd "$PLUGIN_DIR" && nvm install)
}

step_pkg_manager() {
    require_cmd jq
    local pm pm_version
    pm="$(detect_pm)"
    pm_version="$(latest_pm_version "$pm")"
    if [ -z "$pm_version" ]; then
        echo "ERROR: failed to determine latest version for '$pm'" >&2
        exit 1
    fi
    log "Updating packageManager to ${pm}@${pm_version}"

    # Activate via corepack so the binary is available without polluting global npm.
    in_plugin bash -c "
        corepack enable >/dev/null 2>&1 || true
        corepack prepare '${pm}@${pm_version}' --activate
    "

    # Rewrite packageManager field (drop any sha checksum suffix).
    local pkg="$PLUGIN_DIR/package.json"
    local tmp
    tmp="$(mktemp)"
    jq --arg v "${pm}@${pm_version}" '.packageManager = $v' "$pkg" > "$tmp"
    mv "$tmp" "$pkg"
}

step_cp_update() {
    log "Running @grafana/create-plugin update --force"
    in_plugin bash -c "npx --yes @grafana/create-plugin@latest update --force"
}

# Rewrite every non-exact version string in package.json (dependencies,
# devDependencies, peerDependencies, optionalDependencies, overrides,
# resolutions, pnpm.overrides) to an exact resolved version.
step_pin_exact() {
    require_cmd jq
    require_cmd node
    local pkg="$PLUGIN_DIR/package.json"
    log "Pinning all package.json versions to exact"

    ensure_save_exact_npmrc

    # Build a list of "section\tpath\tname\trange" entries to process.
    # Sections handled: dependencies, devDependencies, peerDependencies,
    # optionalDependencies, overrides (npm/yarn shorthand), resolutions (yarn),
    # pnpm.overrides (pnpm).
    local entries
    entries="$(
        jq -r '
            def emit(section; obj):
                (obj // {}) | to_entries[] | "\(section)\t\(.key)\t\(.value)";
            emit("dependencies"; .dependencies),
            emit("devDependencies"; .devDependencies),
            emit("peerDependencies"; .peerDependencies),
            emit("optionalDependencies"; .optionalDependencies),
            emit("overrides"; .overrides),
            emit("resolutions"; .resolutions),
            emit("pnpm.overrides"; (.pnpm // {}).overrides)
        ' "$pkg"
    )"

    if [ -z "$entries" ]; then
        return
    fi

    local need_install=0
    while IFS=$'\t' read -r section name range; do
        [ -z "$section" ] && continue
        # Skip nested object overrides (npm allows nested objects); we only
        # touch leaf string values. jq emits objects as JSON strings starting
        # with "{".
        case "$range" in
            "{"*) continue ;;
        esac
        # Skip non-registry sources we cannot resolve safely.
        case "$range" in
            file:*|link:*|portal:*|workspace:*|git+*|git:*|http:*|https:*|github:*|npm:*)
                continue
                ;;
        esac
        # Already exact? Match X, X.Y, X.Y.Z and prerelease/build metadata.
        if printf '%s' "$range" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+([-+][0-9A-Za-z.+-]+)?$'; then
            continue
        fi

        local resolved
        resolved="$(npm view "${name}@${range}" version 2>/dev/null | tail -1 | awk '{print $2}' | tr -d "'\"")"
        if [ -z "$resolved" ]; then
            # Fall back to the default dist-tag.
            resolved="$(npm_latest "$name")"
        fi
        if [ -z "$resolved" ]; then
            log "  WARN: could not resolve ${name}@${range}; leaving as-is"
            continue
        fi

        # Sanity check: resolved must be exact semver.
        if ! printf '%s' "$resolved" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+([-+][0-9A-Za-z.+-]+)?$'; then
            log "  WARN: resolver returned non-exact value for ${name}: '$resolved'; skipping"
            continue
        fi

        log "  pin: ${section} ${name} ${range} -> ${resolved}"

        local tmp
        tmp="$(mktemp)"
        case "$section" in
            pnpm.overrides)
                jq --arg n "$name" --arg v "$resolved" '.pnpm.overrides[$n] = $v' "$pkg" > "$tmp"
                ;;
            *)
                jq --arg s "$section" --arg n "$name" --arg v "$resolved" '.[$s][$n] = $v' "$pkg" > "$tmp"
                ;;
        esac
        mv "$tmp" "$pkg"
        need_install=1
    done <<< "$entries"

    # Final sanity: assert no non-exact string values remain in scope.
    local bad
    bad="$(
        jq -r '
            def check(section; obj):
                (obj // {}) | to_entries[]
                | select((.value | type) == "string")
                | select(.value | test("^[0-9]+\\.[0-9]+\\.[0-9]+([-+][0-9A-Za-z.+-]+)?$") | not)
                | select(.value | test("^(file:|link:|portal:|workspace:|git\\+|git:|http:|https:|github:|npm:)") | not)
                | "\(section): \(.key) = \(.value)";
            check("dependencies"; .dependencies),
            check("devDependencies"; .devDependencies),
            check("peerDependencies"; .peerDependencies),
            check("optionalDependencies"; .optionalDependencies),
            check("overrides"; .overrides),
            check("resolutions"; .resolutions),
            check("pnpm.overrides"; (.pnpm // {}).overrides)
        ' "$pkg"
    )"
    if [ -n "$bad" ]; then
        log "  WARN: the following entries remain non-exact and could not be resolved:"
        printf '%s\n' "$bad" | sed 's/^/    /' >&2
    fi

    if [ "$need_install" -eq 1 ]; then
        log "Refreshing lockfile after pinning"
        local pm
        pm="$(detect_pm)"
        in_plugin "$pm" install
    fi
}

step_list_grafana() {
    require_cmd jq
    # Emit one "<section>\t<name>\t<current>" per @grafana/* dep for Claude
    # to bump. Sections covered: dependencies, devDependencies.
    jq -r '
        def emit(section; obj):
            (obj // {}) | to_entries[]
            | select(.key | startswith("@grafana/"))
            | "\(section)\t\(.key)\t\(.value)";
        emit("dependencies"; .dependencies),
        emit("devDependencies"; .devDependencies)
    ' "$PLUGIN_DIR/package.json"
}

step_go_version() {
    if [ ! -f "$PLUGIN_DIR/go.mod" ]; then
        log "No go.mod; skipping go-version"
        return
    fi
    require_cmd curl
    require_cmd jq
    local v
    v="$(latest_go_version)"
    log "Setting go directive to $v"
    # Replace the top-level `go X.Y[.Z]` directive.
    local tmp
    tmp="$(mktemp)"
    awk -v ver="$v" '
        !done && /^go [0-9]/ { print "go " ver; done=1; next }
        { print }
    ' "$PLUGIN_DIR/go.mod" > "$tmp"
    mv "$tmp" "$PLUGIN_DIR/go.mod"
}

step_go_sdk() {
    if [ ! -f "$PLUGIN_DIR/go.mod" ]; then
        log "No go.mod; skipping go-sdk"
        return
    fi
    require_cmd go
    log "Bumping grafana-plugin-sdk-go to latest"
    (cd "$PLUGIN_DIR" && go get github.com/grafana/grafana-plugin-sdk-go@latest)
}

step_osv_scan() {
    require_cmd osv-scanner
    log "Running osv-scanner"
    # Try v2 invocation first, fall back to v1.
    if osv-scanner --help 2>&1 | grep -q '^  scan'; then
        (cd "$PLUGIN_DIR" && osv-scanner scan source --format json -r . || true)
    else
        (cd "$PLUGIN_DIR" && osv-scanner --format json -r . || true)
    fi
}

step_go_tidy() {
    if [ ! -f "$PLUGIN_DIR/go.mod" ]; then
        log "No go.mod; skipping go-tidy"
        return
    fi
    require_cmd go
    log "go mod tidy"
    (cd "$PLUGIN_DIR" && go mod tidy)
}

step_verify() {
    require_cmd jq
    local pm
    pm="$(detect_pm)"
    log "Verifying build with $pm"
    in_plugin "$pm" install
    in_plugin "$pm" run typecheck
    in_plugin "$pm" run lint
    in_plugin "$pm" run build
    if [ -f "$PLUGIN_DIR/Magefile.go" ]; then
        require_cmd mage
        log "Building backend via mage"
        (cd "$PLUGIN_DIR" && go mod download -x && mage -v buildAll)
    fi
}

# --- dispatch -----------------------------------------------------------------

run_step() {
    case "$1" in
        node-version)  step_node_version ;;
        pkg-manager)   step_pkg_manager ;;
        cp-update)     step_cp_update ;;
        pin-exact)     step_pin_exact ;;
        list-grafana)  step_list_grafana ;;
        go-version)    step_go_version ;;
        go-sdk)        step_go_sdk ;;
        osv-scan)      step_osv_scan ;;
        go-tidy)       step_go_tidy ;;
        verify)        step_verify ;;
        *)
            echo "ERROR: unknown step: $1" >&2
            exit 1
            ;;
    esac
}

if [ "$RUN_ALL" -eq 1 ]; then
    # Order matches the slash command's orchestration. list-grafana and
    # osv-scan are informational steps Claude consumes; --all still runs them
    # so the output is captured for inspection but the real remediation has
    # to be driven by Claude.
    for s in \
        node-version \
        pkg-manager \
        cp-update \
        pin-exact \
        list-grafana \
        go-version \
        go-sdk \
        osv-scan \
        go-tidy \
        pin-exact \
        verify
    do
        run_step "$s"
    done
else
    run_step "$STEP"
fi
