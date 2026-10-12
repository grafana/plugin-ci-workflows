# Publish Playwright Reports Action

Uploads Playwright test reports to Google Cloud Storage and comments on the pull request with results and links. Designed to work together with the `upload-report-artifacts` action.

> [!IMPORTANT]
> Do not call this action directly from your own workflow. It cannot authenticate there.
> Call the [`playwright-publish-reports.yml`](../../../../.github/workflows/playwright-publish-reports.yml) reusable workflow instead, as shown under [Usage](#usage).
> The reason is under [Requirements](#requirements).

Reports are stored at:
```
gs://grafana-e2e-test-artifacts/{owner}/{repo}/{run-id}/{run-attempt}/{matrix-dir}/
```

Each run attempt gets its own prefix, so a re-run never overwrites a report that an earlier comment or job summary links to.

Report links in PR comments require a Grafana Google Workspace sign-in and are retained for 30 days via a GCS object lifecycle rule.

## Usage

Plugins that use `ci.yml` with `upload-playwright-artifacts: true` get this automatically. Nothing to configure.

Repositories with their own Playwright workflow upload the reports with `upload-report-artifacts`, then call the reusable workflow in a separate job:

```yaml
jobs:
  playwright-tests:
    # ...
    steps:
      # ... run the tests with id: run-tests, then:
      - uses: grafana/plugin-ci-workflows/actions/plugins/playwright-reports/upload-report-artifacts@plugins-playwright-reports-upload-report-artifacts/vX.Y.Z
        if: ${{ always() && (steps.run-tests.outcome == 'success' || steps.run-tests.outcome == 'failure') }}
        with:
          test-outcome: ${{ steps.run-tests.outcome }}

  publish-reports:
    needs: [playwright-tests]
    if: ${{ always() && !cancelled() && (github.event_name != 'pull_request' || github.event.pull_request.head.repo.fork == false) }}
    uses: grafana/plugin-ci-workflows/.github/workflows/playwright-publish-reports.yml@ci-cd-workflows/vX.Y.Z
    permissions:
      id-token: write
      pull-requests: write
      contents: read
```

- **Reference the workflow by a release tag or `@main`.** A commit SHA does not authenticate. The service account trusts `job_workflow_ref` at `refs/heads/main` or at a semver release tag, and a SHA matches neither.
- **Grant the permissions on the calling job.** A called workflow cannot exceed the permissions of its caller.
- **Keep the fork guard.** GitHub withholds `id-token: write` from fork pull requests, so the job would fail on them.

## Inputs

| Input Name           | Description                                                      | Required | Default                  |
| -------------------- | ---------------------------------------------------------------- | -------- | ------------------------ |
| `pr-comment-summary` | Whether to post a PR comment with test results and report links. | Yes      | `true`                   |
| `artifact-pattern`   | Pattern to match the uploaded artifacts.                         | Yes      | `gf-playwright-report-*` |

The bucket and the service account are intentionally not inputs. They are shared internal Grafana resources with a single environment, so callers do not need to know or configure them.

The reusable workflow exposes `pr-comment-summary` only.

## Requirements

Uploads impersonate `github-e2e-test-artifacts@grafanalabs-workload-identity.iam.gserviceaccount.com`, which holds `roles/storage.objectUser` on the `grafana-e2e-test-artifacts` bucket (project `grafanalabs-global`).

The service account trusts exactly one workflow file, `plugin-ci-workflows/.github/workflows/playwright-publish-reports.yml`, through the `job_workflow_ref` attribute at `main` and at release tags. So:

- Any repository that calls that workflow inherits access. It can call it directly, or through `ci.yml` and `playwright.yml`. There is no `deployment_tools` onboarding step.
- Every event works, including `push` and `schedule`. The attribute does not constrain the event.
- A workflow that another repository writes itself cannot obtain the credential. That includes one that uses this action directly: the job then runs under that repository's workflow file, so `job_workflow_ref` names that file and does not match.

Read access to the bucket is granted to `domain:grafana.com`, which is why report links require a Grafana Google Workspace sign-in. Objects are deleted after 30 days by the bucket's lifecycle rule.

### Why the trusted unit is a file, not a job

GitHub's `job_workflow_ref` claim names the workflow **file** that defines a job. There is no claim for the job itself. Any job in a trusted file can therefore present the trusted identity.

Before, these steps lived in a job inside `playwright.yml`. That file also defines `playwright-tests`, which checks out the caller and runs its dependency install and its tests. Binding the service account to `playwright.yml` would hand the credential to that caller code as well. A separate job does not change the claim. A separate file does, which is why `playwright-publish-reports.yml` exists.

### Per-repository isolation is enforced here, not by IAM

The service account can write anywhere in the bucket. Each repository stays inside its own `<owner>/<repository>/` prefix because this action builds the destination from `github.repository`, which a caller cannot set. IAM does not enforce it, so the following must stay true:

- No input that a caller controls may reach the upload destination.
- `playwright-publish-reports.yml` must not check out the caller's repository.
- No step may execute content from the downloaded artifacts. Reading them as data is fine.
- The publish steps must stay in `playwright-publish-reports.yml`. Inlining them into `playwright.yml`, or into any other workflow that runs caller code, moves the credential next to that code.

Breaking any of these would let one repository write to another repository's prefix, with no IAM backstop. The same rules are recorded at the top of `playwright-publish-reports.yml`.

## Upload performance

A failing Playwright report is thousands of small files. Two things keep the upload quick:

- Reports are uploaded with `resumable: false`. A resumable upload spends an extra round trip per file on the initiation request.
- `upload-report-artifacts` strips the report body for passing matrix legs, so only failures carry a full report. Leave `upload-successful-reports` at its default unless you specifically need passing reports published.

If the upload starts running long, the `concurrency` input of `google-github-actions/upload-cloud-storage` (default `100`) is the next knob to reach for.
