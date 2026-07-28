# Act testing guide

`tests/act` exercises plugin CI/CD workflows with [nektos/act](https://github.com/nektos/act), Go, and Docker. Use it for behavior that depends on workflow orchestration. The framework builds temporary, narrowed workflow copies, runs them in act containers, and records workflow results for assertions.

## Policy and selection

For behavioral changes to plugin-facing reusable CI/CD workflows, published plugin actions, and the internal actions that support them, add or update the narrowest credible automated coverage whenever possible. Existing coverage can be sufficient. This policy does not impose a general coverage requirement on repository release or maintenance automation.

Choose the smallest suitable level:

1. **Static or unit coverage** for pure wiring: input declarations and pass-throughs, references, paths, and other behavior that does not require a running workflow. See `main_workflow_consistency_test.go` for examples.
2. **Direct action coverage** when a small test workflow can call the action and assert its contract without going through the CI/CD wrappers.
3. **CI wrapper coverage** for behavior in `.github/workflows/ci.yml`, using `internal/workflow/ci` helpers.
4. **CD wrapper coverage** when the behavior crosses `.github/workflows/cd.yml` and CI, using `internal/workflow/cd` helpers.

Do not duplicate adequate existing coverage. If act cannot faithfully emulate the behavior, document the precise limitation and the alternative validation in the PR description.

## Prerequisites and commands

Run commands from the repository root. Workflow execution requires a working Docker daemon, the `act` CLI, Go, and either `GITHUB_TOKEN` or an authenticated `gh` CLI (`gh auth token`). `make act-lint` additionally requires `golangci-lint`.

Act and `actions/checkout` clone the commit under test, so run act tests from a checkout whose committed HEAD is pushed and remotely available. A committed but unpushed change causes the clone to fail. Uncommitted working-tree changes are acceptable and do not cause this issue.

```bash
make actionlint
make act-lint
make act-test

# Focus on one test while iterating
cd tests/act && go test -v -timeout 1h -run '^TestGCS$'
```

`make act-test` runs `go test -v -timeout 1h` in `tests/act`. It may create temporary `act-*.yml` workflows and use temporary directories under `/tmp`. `make clean` removes act temporary data and cached action data.

When a fixture plugin under `tests/simple-*` changes, regenerate its pre-built artifacts with `make mockdata`.

## Framework capabilities

The runner in `internal/act/act.go` maps this repository's workflow and action references to the local checkout, maps supported self-hosted runner labels to act's runner image, and exposes results through `RunResult`:

- workflow success, step outputs, summaries, and GitHub Actions annotations;
- uploaded artifacts through `runner.ArtifactsStorage`;
- a local, read-only view of mocked GCS uploads through `runner.GCS.Fs`.

The testing workflow model in `internal/workflow/testing.go` supports narrowed execution and mutations. It can keep one job, remove jobs, no-op or replace steps, stop after a step, provide dependency outputs, inject steps, override matrices, and alter step environments or workflow triggers. CI and CD wrapper builders are in `internal/workflow/ci/ci.go` and `internal/workflow/cd/cd.go`; they create nested temporary workflows so the called CI/CD workflows can be mutated.

Event helpers in `internal/act/workflows.go` cover push, pull request (including fork and actor variants), pull-request-target, release, and workflow-dispatch payloads. Tests can use custom payloads when those helpers do not contain the required event fields.

### Mocked boundaries

Mock external boundaries, not the behavior under test. The checked-in helpers support:

- GCS authentication/upload replacement with local file assertions;
- Vault-secret and GitHub App-token step replacements;
- GCOM HTTP handlers that receive requests from act containers;
- Argo-workflow trigger replacement backed by an HTTP request spy;
- generic Docker-accessible HTTP spies for request/input assertions;
- pre-built plugin `dist` and packaged artifact fixtures through `tests/act/mockdata` and CI wrapper helpers.

Use the mock that preserves the changed contract. For example, mock a GCS service while asserting the paths emitted by the workflow, rather than no-oping the upload step when upload-path behavior is under test.

## Go testing conventions

Use `github.com/stretchr/testify/require` rather than standard-library assertions:

```go
func TestSomething(t *testing.T) {
	result, err := DoSomething()
	require.NoError(t, err)
	require.Equal(t, expected, result)
}
```

Prefer table-driven tests, and run independent subtests in parallel:

```go
for _, tc := range []testCase{
	{name: "case1", input: "a", expected: "A"},
	{name: "case2", input: "b", expected: "B"},
} {
	t.Run(tc.name, func(t *testing.T) {
		t.Parallel()
		require.Equal(t, tc.expected, Transform(tc.input))
	})
}
```

## Writing focused tests

Start from a representative test and keep the executable slice small:

- `main_smoke_test.go` runs the basic CI workflow.
- `main_cd_test.go` exercises complex CD behavior with mocks and workflow mutations.
- `main_backend_build_target_test.go` demonstrates assertions through `::act-debug::` annotations.
- `main_context_test.go` narrows CI to the workflow-context step and varies event payloads.
- `main_gcs_test.go` uses mocked plugin artifacts and GCS, then asserts uploaded files and outputs.
- `main_cd_argo_test.go` asserts the mocked Argo request contract.
- `main_package_test.go` and `main_smoke_test.go` inspect uploaded artifacts.
- `main_workflow_consistency_test.go` performs static workflow and release-configuration consistency checks.
- `internal/workflow/ci/ci.go` provides CI workflow helpers.
- `internal/act/act.go` implements the runner.

Prefer assertions on outputs, artifacts, mocked files, summaries, and recorded HTTP requests. The runner also captures annotations.

### Debug annotations

Use the custom `::act-debug::` annotation only when an observable contract cannot otherwise express the assertion. The framework stores it as an `AnnotationLevelDebug` annotation.

Emit an annotation from a workflow step with logfmt:

```yaml
- run: printf '::act-debug::msg="%s" key=%s\n' "my step ran" "${SOME_VAR}"
  env:
    SOME_VAR: ${{ inputs.something }}
```

Or emit one from a Mage target:

```go
fmt.Printf("::act-debug::msg=%q\n", "my target was invoked")
```

Assert it in the test:

```go
expMsg, err := logfmt.MarshalKeyvals("msg", "my target was invoked")
require.NoError(t, err)
require.Contains(t, r.Annotations, act.Annotation{
	Level:   act.AnnotationLevelDebug,
	Message: string(expMsg),
})
```

### Keep act tests focused and fast

Run only the job and steps that establish the behavior. Scope CI tests with `MutateCIWorkflow().With(...)`:

```go
ci.MutateCIWorkflow().With(
	// Keep only the job under test; strip all other jobs and clear its `needs`.
	workflow.WithOnlyOneJob(t, "test-and-build", true),
	// No-op steps that are irrelevant to the behavior under test.
	workflow.WithNoOpStep(t, "test-and-build", "frontend"),
	// Drop packaging, signing, upload, and later steps.
	workflow.WithRemoveAllStepsAfter(t, "test-and-build", "backend"),
)
```

- `WithOnlyOneJob(t, jobID, removeDependencies)` removes all jobs except `jobID`. Pass `true` to clear its `needs` so it can run standalone.
- `WithNoOpStep(t, jobID, stepID)` replaces a step with the shell no-op `:` while retaining its ID and output wiring.
- `WithRemoveAllStepsAfter(t, jobID, stepID)` drops each following step in the job.

Combine these mutations when only one behavior matters. Reuse mockdata rather than rebuilding fixture plugins, and keep independent Go subtests parallel when their resources are isolated.

The repository is mounted into act's Docker container, so large local trees slow test startup. In particular, `node_modules` directories in fixture plugins under `tests/`, such as `tests/simple-frontend` and `tests/simple-backend`, add many files. Before running act tests when they exist, run `make clean`; its `clean-node-modules` target removes `node_modules` recursively beneath `tests/`.

Add CI/CD wrapper helper inputs only when a test needs to set that input. The helpers do not need to mirror every workflow input merely for parity.

## Limits and alternatives

Act runs some Docker-based behavior successfully in this repository, including workflow steps that use the framework's mounted mockdata and Docker-accessible HTTP mocks. Docker use alone is not a limitation.

However, act is not full GitHub Actions infrastructure. Concrete limitations can include job containers, Docker-in-Docker or Docker Compose, browser infrastructure, OIDC/workload identity federation, and fidelity to GitHub-hosted runners. Dynamic matrix behavior also has framework workarounds rather than faithful act execution. When one of these prevents credible coverage, state the exact unsupported behavior in the PR description and record the alternative validation performed, such as a targeted manual run in the appropriate environment or a static check of the wiring.
