# Publish Playwright Reports Action

Uploads Playwright test reports to Google Cloud Storage and comments on the pull request with results and links. Designed to work together with the `upload-report-artifacts` action.

Reports are stored at:
```
gs://grafana-e2e-test-artifacts/{owner}/{repo}/{YYYYMMDD}/{pr-number-or-run-id}/{matrix-dir}/
```

Report links in PR comments require a Grafana Google Workspace sign-in and are retained for 90 days via a GCS object lifecycle rule.

## Inputs

| Input Name           | Description                                                      | Required | Default                  |
| -------------------- | ---------------------------------------------------------------- | -------- | ------------------------ |
| `pr-comment-summary` | Whether to post a PR comment with test results and report links. | Yes      | `true`                   |
| `artifact-pattern`   | Pattern to match the uploaded artifacts.                         | Yes      | `gf-playwright-report-*` |

The bucket and the service account are intentionally not inputs. They are shared internal Grafana resources with a single environment, so callers do not need to know or configure them.

## Requirements

The calling job needs the following permissions:

```yaml
permissions:
  id-token: write # to authenticate to GCS
  pull-requests: write # to post the summary comment
```

Uploads impersonate `github-e2e-test-artifacts@grafanalabs-workload-identity.iam.gserviceaccount.com`, which holds `roles/storage.objectUser` on the `grafana-e2e-test-artifacts` bucket (project `grafanalabs-global`).

That service account is impersonable only from this repository's Playwright workflows, via the `job_workflow_ref` attribute pinned to `main` and to release tags. This is the same trust pattern as `github-plugin-ci-workflows@`, so:

- Any repository calling `playwright.yml` inherits access. There is no `deployment_tools` onboarding step.
- Every event works, including `push` and `schedule`. The attribute does not constrain the event.
- A workflow that a plugin repository writes itself cannot obtain the credential. Only this repository's pinned workflows can.

Read access to the bucket is granted to `domain:grafana.com`, which is why report links require a Grafana Google Workspace sign-in. Objects are deleted after 90 days by the bucket's lifecycle rule.

### Per-repository isolation is enforced here, not by IAM

The service account can write anywhere in the bucket. Each repository stays inside its own `<owner>/<repository>/` prefix because this action builds the destination from `github.repository`, which a caller cannot set. IAM does not enforce it, so the following must stay true:

- No input that a caller controls may reach the upload destination.
- The publish job must not check out the caller's repository.
- No step may execute content from the downloaded artifacts. Reading them as data is fine.
- The upload must stay in its own job, separate from the job that runs the caller's tests.

Breaking any of these would let one repository write to another repository's prefix, with no IAM backstop.

## Upload performance

The federated credentials this action uses are short-lived, so the upload has to finish promptly. Two things keep it inside that window:

- Reports are uploaded with `resumable: false`. A failing Playwright report is thousands of small files, and a resumable upload spends an extra round trip per file on the initiation request.
- `upload-report-artifacts` strips the report body for passing matrix legs, so only failures carry a full report. Leave `upload-successful-reports` at its default unless you specifically need passing reports published.

If the upload starts running long, the `concurrency` input of `google-github-actions/upload-cloud-storage` (default `100`) is the next knob to reach for.
