---
description: Refresh deps and tooling across all test plugins in tests/ (excludes tests/act/)
argument-hint: "[plugin-name]"
allowed-tools: Bash, Read, Edit, Write, Glob, Grep, TodoWrite
---

You are performing a chore refresh of one or more Grafana test plugins under
`tests/`. The folder `tests/act/` is the Go test harness, NOT a plugin —
never touch it. Other folders directly under `tests/` are plugins to update.

A helper script `scripts/update-test-plugin.sh` performs the deterministic
work. Your job is to orchestrate it, apply judgment to the non-deterministic
parts (`@grafana/*` bumps, vuln triage, build failures), enforce the exact-pin
invariant, and produce a final summary.

# Hard invariants

1. **Every version string in any plugin's `package.json` MUST be exact semver**
   (e.g. `1.2.3`, `1.2.3-rc.1`). No `^`, `~`, `>`, `<`, `>=`, `<=`, `*`, `x`,
   `||`, or whitespace ranges. This applies to `dependencies`,
   `devDependencies`, `peerDependencies`, `optionalDependencies`, `overrides`,
   `resolutions`, and `pnpm.overrides`. Non-registry specifiers (`file:`,
   `link:`, `workspace:`, `git+`, `http(s):`, `github:`, `npm:`) are allowed
   as-is.
2. **Never `git add`, `git commit`, or push**. Leave changes unstaged.
3. **Never modify `tests/act/`** in any way.

# Step inventory (per plugin)

Run via `./scripts/update-test-plugin.sh <plugin> --step <step>`:

| Step           | Purpose                                                            |
|----------------|--------------------------------------------------------------------|
| `node-version` | Write latest LTS major to `.nvmrc`, install via nvm                |
| `pkg-manager`  | Update `packageManager` field; activate via Corepack (yarn stays 1.x) |
| `cp-update`    | `npx @grafana/create-plugin@latest update --force`                 |
| `pin-exact`    | Rewrite every non-exact dep in package.json to exact + reinstall   |
| `list-grafana` | Print `<section>\t<name>\t<current>` for every `@grafana/*` dep    |
| `go-version`   | Bump `go` directive in `go.mod` to latest stable (skips if no go.mod) |
| `go-sdk`       | `go get github.com/grafana/grafana-plugin-sdk-go@latest`           |
| `osv-scan`     | Run `osv-scanner` and emit JSON                                    |
| `go-tidy`      | `go mod tidy`                                                      |
| `verify`       | `<pm> install`, typecheck, lint, build, plus mage if `Magefile.go` |

# Workflow

1. **Pick plugins.**
   - If `$ARGUMENTS` is non-empty, treat it as a single plugin name. Validate
     it appears in `./scripts/find-tests.sh`. If not, error out and stop.
   - Otherwise, iterate over every line printed by
     `./scripts/find-tests.sh` (each is `./<plugin>`; strip the `./`). Skip
     any entry whose name starts with `.` (e.g. `.claude`) — those are not
     plugins. The helper script also enforces this defensively.

2. **Create todos.** Use `TodoWrite` to track work as (plugin, phase) pairs.

3. **Per plugin, run in this order** (continue to the next plugin on failure,
   but record the failure for the final summary):

   1. `node-version`
   2. `pkg-manager`
   3. `cp-update`
   4. `pin-exact` (pass 1 — clean up ranges create-plugin reintroduced)
   5. **Bump `@grafana/*` packages**:
      - Run `--step list-grafana` to get the package list.
      - For each `<name>`, resolve the exact latest version:
        `npm view <name> version`.
      - Install all bumps in a single command for the plugin's package manager,
        always with the exact-version flag:
        - npm:  `npm install <pkg>@<exact> <pkg>@<exact> ... --save-exact`
        - pnpm: `pnpm add  <pkg>@<exact> <pkg>@<exact> ... --save-exact`
        - yarn: `yarn add  <pkg>@<exact> <pkg>@<exact> ... --exact`
      - The package manager / install must run with the plugin's Node active.
        Use: `cd tests/<plugin> && source ~/.nvm/nvm.sh && nvm use && <pm> ...`
      - Respect the section: dependencies bumps use the default flag,
        devDependencies bumps add `-D` (npm/pnpm) or `--dev` (yarn).
   6. `go-version` (skipped automatically if no go.mod)
   7. `go-sdk`     (skipped automatically if no go.mod)
   8. **`osv-scan` and remediate**:
      - Run `--step osv-scan` and parse the JSON output.
      - For each vulnerability with severity `HIGH` or `CRITICAL`:
        - Identify the affected package and ecosystem (npm / Go).
        - If the affected package is a direct dependency, bump it to an exact
          fixed version via `<pm> add ... --save-exact` (npm) or `go get`.
        - If transitive (npm), add an entry to the appropriate exact-version
          override map:
          - npm  → top-level `overrides`
          - yarn → top-level `resolutions`
          - pnpm → `pnpm.overrides`
          Always use a single exact version, never a range.
        - If transitive (Go), bump the offending direct dep that pulls it in,
          or add a `require <module> <version>` line for the transitive module
          to force the patched version, then `go mod tidy`.
      - After remediation, re-run `osv-scan` and confirm no `HIGH`/`CRITICAL`
        vulns remain. If any remain that cannot be cleanly remediated,
        record them in the summary and move on.
   9. `go-tidy`
   10. `pin-exact` (pass 2 — final invariant sweep)
   11. `verify`

4. **On `verify` failure — best-effort iterative fix.** Try multiple targeted
   fixes until verify passes or you run out of plausible options. Examples:
   - Re-read the error output, identify the failing package, downgrade just
     that one to a known-working exact version.
   - Re-run `cp-update` if config files appear inconsistent.
   - Re-run `pin-exact` if any range slipped back in.
   - Clean `node_modules` and lockfile fragments, then re-install.
   - For mage failures, ensure `go-tidy` was run after all Go bumps.
   After each attempted fix, re-run `verify`. If still failing after several
   attempts, mark the plugin failed and continue.

5. **Final invariant check.** Before reporting a plugin as green, grep its
   `package.json` to confirm no value matches a non-exact pattern. Quick
   sanity check:

   ```
   jq -r '
     [.dependencies, .devDependencies, .peerDependencies, .optionalDependencies,
      .overrides, .resolutions, (.pnpm // {}).overrides]
     | map(. // {}) | map(to_entries[])
     | .[] | select((.value | type) == "string")
     | select(.value | test("^(file:|link:|portal:|workspace:|git\\+|git:|http:|https:|github:|npm:)") | not)
     | select(.value | test("^[0-9]+\\.[0-9]+\\.[0-9]+([-+][0-9A-Za-z.+-]+)?$") | not)
     | "\(.key) = \(.value)"
   ' tests/<plugin>/package.json
   ```

   If anything is printed, fix it and re-run `pin-exact` + `verify`.

6. **Summary.** Print a markdown table at the end:

   ```
   | plugin | node | pm | cp-update | pin1 | @grafana | go | osv | tidy | pin2 | verify |
   |--------|------|----|-----------|------|----------|----|-----|------|------|--------|
   ```

   Use ✓ / ✗ / — (skipped/n-a) per cell. Below the table, for each ✗ cell
   include a one-line cause.

7. **Mockdata.** If every plugin reports green across the board, run
   `make mockdata` to refresh `tests/act/mockdata/dist/*`. If anything failed,
   skip mockdata and explain that the user should re-run after addressing the
   failures.

8. **Done.** Do not stage or commit. Just print the summary and stop.

# Tips

- The helper validates the plugin argument against `find-tests.sh` so you
  cannot accidentally point it at `tests/act/`.
- All shell commands the helper issues run with `nvm use` inside the plugin
  dir, so the right Node is active automatically.
- `pin-exact` is idempotent — safe to run repeatedly.
- Yarn 1.x is intentional (v4 migration is pending). Do NOT bump yarn to v2+.
- The plugin sources are dummy; you do not need to read or modify any plugin
  source code. If a build fails, it is almost always a dependency issue.
