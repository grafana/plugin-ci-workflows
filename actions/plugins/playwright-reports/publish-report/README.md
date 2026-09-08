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

The bucket and the service account used for the upload are intentionally not inputs — they are shared internal Grafana resources with a single environment, so callers don't need to know or configure them.

## Requirements

The calling job needs the following permissions:

```yaml
permissions:
  id-token: write # to authenticate to GCS
  pull-requests: write # to post the summary comment
```

Uploads run as `github-e2e-test-artifacts@grafanalabs-workload-identity.iam.gserviceaccount.com`, which holds `roles/storage.objectUser` on the `grafana-e2e-test-artifacts` bucket (project `grafanalabs-global`).

**Impersonating that service account is default-deny per repository.** To onboard a new repository, add an entry to the `e2e-test-artifacts` list in `deployment_tools`, in `terraform/projects/grafanalabs-workload-identity/github-oidc-service-accounts.tf`. At the time of writing, the allowlist covers only `grafana-test-datasource` and `plugin-tools`, and only for `pull_request` events on `branch` refs — pushes and tags are not authorized yet. A repository that is not on the allowlist will fail at the "Login to GCS" step.

Read access to the bucket is granted to `domain:grafana.com`, which is why report links require a Grafana Google Workspace sign-in. Objects are deleted after 90 days by the bucket's lifecycle rule.
