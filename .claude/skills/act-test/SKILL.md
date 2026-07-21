---
name: act-test
description: Add or update focused automated coverage for plugin CI/CD workflow and action behavior
---

# Act testing

Use this skill for behavioral changes to plugin-facing reusable CI/CD workflows, published plugin actions, and their supporting internal actions. Read the [act testing guide](../../../tests/act/CLAUDE.md) before changing tests.

## Steps

1. Identify the observable behavioral contract: outputs, artifacts, external request payloads, generated files, or workflow success/failure.
2. Select the narrowest credible coverage: static/unit coverage for pure wiring, a direct action test workflow for action-local behavior, a CI wrapper test, or a CD wrapper test. Existing coverage may already be adequate.
3. Isolate the smallest executable slice. Use `tests/act/internal/workflow/testing.go` mutations to retain only the relevant jobs, dependencies, and steps. Use the CI and CD wrapper helpers in `tests/act/internal/workflow/ci/ci.go` and `tests/act/internal/workflow/cd/cd.go` when testing the reusable wrappers.
4. Mock external boundaries, not the behavior under test. Use the existing GCS, Vault, GitHub App token, GCOM, Argo, HTTP-spy, artifact, and mockdata helpers in `tests/act/internal/act/` and `tests/act/internal/workflow/` as appropriate.
5. Assert outputs, artifacts, mocked files, summaries, or recorded HTTP requests before using `::act-debug::` annotations. Use annotations only when the observable contract cannot otherwise be asserted.
6. If act cannot faithfully emulate the behavior, document the precise limitation and the alternative validation in the PR description. Docker use alone is not a limitation.
7. Run the focused test first, then the broader relevant checks such as `make act-test`, `make act-lint`, or `make actionlint`.

Do not add helper inputs merely to mirror workflow inputs; add them when a test needs to set them.

Before local act execution, ensure the committed HEAD is pushed and remotely available; uncommitted working-tree changes are fine. Run `make clean` first when fixture-plugin `node_modules` directories exist. See the guide for details.
