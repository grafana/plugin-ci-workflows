# Code Review Guidelines for plugin-ci-workflows

You are reviewing code in the `grafana/plugin-ci-workflows` repository. This repository contains reusable GitHub Actions workflows and actions for Grafana plugin CI/CD.

## Project Structure Awareness

### Workflows (`.github/workflows/`)
- `ci.yml` and `cd.yml` are **reusable workflows** for plugin users
- `pr-checks-*` and `release-please-*` are internal repo workflows
- Everything else (like `playwright.yml`) are internal to one of the reusable workflows — not for direct user use

### Actions (`actions/`)
- `actions/internal/` — Actions coupled to workflow internals, NOT for direct user use
- `actions/plugins/` — Standalone reusable actions for users

### Examples (`examples/`)
- `examples/base/` — Core workflow examples with auto-generated README
- `examples/extra/` — Additional helper examples
- **`examples/base/README.md` is auto-generated** by `examples/base/genreadme.go` — never edit directly

### Tests (`tests/`)
- `tests/act/` — Go testing framework using nektos/act in Docker to run workflows
- Every other folder in `tests` (like `tests/simple-frontend`) — Dummy Grafana plugins used as test fixtures
- `tests/act/mockdata/` — Pre-generated artifacts (built plugins, ZIPs) for fast test execution
- **The testing framework is WIP** — do not write new tests unless explicitly requested

---

## GitHub Actions Security (CRITICAL)

### 1. External Actions Must Be Pinned to Commit SHAs

Flag any external action not pinned to a full commit SHA:

```yaml
# ❌ FAIL - vulnerable to supply chain attacks
uses: actions/checkout@v4
uses: actions/checkout@v4.2.2

# ✅ PASS
uses: actions/checkout@8e8c483db84b4bee98b60c0593521ed34d9990e8 # v4.2.2
```

### 2. No GitHub Expressions in Shell Scripts

Flag any `${{ }}` expressions inside `run:` blocks — this is a shell injection vulnerability:

```yaml
# ❌ FAIL - shell injection vulnerability
- run: |
    echo "Branch: ${{ github.head_ref }}"
    if [ "${{ inputs.environment }}" = "prod" ]; then
      deploy
    fi

# ✅ PASS - use environment variables
- run: |
    echo "Branch: ${BRANCH}"
    if [ "${ENVIRONMENT}" = "prod" ]; then
      deploy
    fi
  env:
    BRANCH: ${{ github.head_ref }}
    ENVIRONMENT: ${{ inputs.environment }}
```

### 3. No `ubuntu-latest` Runner

Flag use of `ubuntu-latest`. Must use self-hosted runner labels:

```yaml
# ❌ FAIL
runs-on: ubuntu-latest

# ✅ PASS - prefer ARM runners (cheaper)
runs-on: ubuntu-arm64-small   # Simple scripts
runs-on: ubuntu-arm64         # Standard workloads
runs-on: ubuntu-arm64-large   # Heavy workloads
```

### 4. Same-Repo References Must Use `@main`

When referencing workflows/actions from this repository:

```yaml
# ❌ FAIL
uses: grafana/plugin-ci-workflows/.github/workflows/ci.yml@v1.0.0

# ✅ PASS
uses: grafana/plugin-ci-workflows/.github/workflows/ci.yml@main
```

### 5. Complex Logic Should Use JavaScript

For complex conditionals or data manipulation, prefer `actions/github-script` over bash:

```yaml
# ✅ PASS - complex logic in JavaScript
- uses: actions/github-script@ed597411d8f924073f98dfc5c65a23a2325f34cd # v8.0.0
  with:
    script: |
      const env = process.env.ENVIRONMENT;
      const allowed = ['dev', 'ops', 'prod'];
      if (!allowed.includes(env)) {
        core.setFailed(`Invalid environment: ${env}`);
      }
  env:
    ENVIRONMENT: ${{ inputs.environment }}
```

---

## Project-Specific Rules

### New Shared Workflows (internal to ci/cd)

If a new workflow file is added that's part of `ci.yml`/`cd.yml` (like `playwright.yml`), verify it's added to the `switch-references` step in ALL THREE files:
- `.github/workflows/pr-checks-workflow-references.yml`
- `.github/workflows/release-please-pr-update-tagged-references.yml`
- `.github/workflows/release-please-restore-rolling-release.yml`

### New User-Facing Actions

If an action is added to `actions/plugins/`, verify `release-please-config.json` is updated.

### Examples Directory

If files in `examples/base/**/*` are modified, remind that `examples/base/README.md` is auto-generated and `make genreadme` should be run.

### Test Plugins

If files in `tests/simple-*` (dummy test plugins) are modified, remind that `make mockdata` should be run to regenerate mock data.

---

## Go Code (tests/act/)

When reviewing Go code in `tests/act/`:

### Error Handling

```go
// ❌ FAIL - bare error return
if err := doSomething(); err != nil {
    return err
}

// ✅ PASS - wrap with context
if err := doSomething(); err != nil {
    return fmt.Errorf("do something: %w", err)
}
```

### Testing

- Use `github.com/stretchr/testify/require` for assertions
- Prefer table-driven tests
- Use `t.Parallel()` for independent tests
- Use `t.Cleanup()` for resource cleanup

### Linting

Remind to run `make act-lint` before merging Go changes.

---

## Review Checklist

For every PR, verify:

- [ ] External actions pinned to commit SHAs with version comments
- [ ] No `${{ }}` expressions in shell scripts
- [ ] No `ubuntu-latest` — using self-hosted runner labels
- [ ] Same-repo references use `@main`
- [ ] New shared workflows added to switch-references in all 3 files
- [ ] New `actions/plugins/` actions added to `release-please-config.json`
- [ ] `make genreadme` run if `examples/base/` modified
- [ ] `make mockdata` run if `tests/simple-*` modified
- [ ] Go errors wrapped with context
- [ ] `make act-lint` and `make actionlint` pass

