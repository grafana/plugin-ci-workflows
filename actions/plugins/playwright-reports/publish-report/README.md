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

The bucket used for the upload is intentionally not an input — it is a shared internal Grafana resource with a single environment, so callers don't need to know or configure it.

## Requirements

The calling job needs the following permissions:

```yaml
permissions:
  id-token: write # to authenticate to GCS
  pull-requests: write # to post the summary comment
```

Uploads authenticate as the repository's **Direct WIF principal set** — there is no service account. The `grafana-e2e-test-artifacts` bucket (project `grafanalabs-global`) grants `roles/storage.objectUser` per repository, under an IAM condition confining each repository to its own `<owner>/<repo>/` object prefix. This is why no service account is used: an IAM condition can only read `resource.*` and `request.*`, so it cannot tell which repository is behind an impersonated service account, which would leave the prefix unenforced.

**Publishing is default-deny per repository.** To onboard a new repository, add an entry to `e2e_artifacts_publishers` in `deployment_tools`, in `terraform/storage/grafanalabs-global/e2e-test-artifacts.tf`. The map key must equal `github.repository` (for example `grafana/plugin-tools`), because it is both the enforced object prefix and the prefix this action uploads to. At the time of writing the allowlist covers only `grafana/grafana-test-datasource` and `grafana/plugin-tools`, and only for `pull_request` events on `branch` refs — pushes and tags are not authorized. A repository that is not on the allowlist fails at the upload step.

Read access to the bucket is granted to `domain:grafana.com`, which is why report links require a Grafana Google Workspace sign-in. Objects are deleted after 90 days by the bucket's lifecycle rule.

## Upload performance

The federated credentials this action uses are short-lived, so the upload has to finish promptly. Two things keep it inside that window:

- Reports are uploaded with `resumable: false`. A failing Playwright report is thousands of small files, and a resumable upload spends an extra round trip per file on the initiation request.
- `upload-report-artifacts` strips the report body for passing matrix legs, so only failures carry a full report. Leave `upload-successful-reports` at its default unless you specifically need passing reports published.

If the upload starts running long, the `concurrency` input of `google-github-actions/upload-cloud-storage` (default `100`) is the next knob to reach for.
