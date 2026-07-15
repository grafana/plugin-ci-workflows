# Publish Playwright Reports Action

Uploads Playwright test reports to Google Cloud Storage and comments on the pull request with results and links. Designed to work together with the `upload-report-artifacts` action.

Reports are stored at:
```
gs://{bucket}/{owner}/{repo}/{YYYYMMDD}/{pr-number-or-run-id}/{matrix-dir}/
```

Report links in PR comments require a Grafana Google Workspace sign-in and are retained for 90 days via a GCS object lifecycle rule.

## Inputs

| Input Name           | Description                                                                                              | Required | Default                          |
| -------------------- | -------------------------------------------------------------------------------------------------------- | -------- | -------------------------------- |
| `pr-comment-summary` | Whether to post a PR comment with test results and report links.                                         | Yes      | `true`                           |
| `artifact-pattern`   | Pattern to match the uploaded artifacts.                                                                 | Yes      | `gf-playwright-report-*`         |
| `bucket`             | GCS bucket name to upload reports to. Defaults to the shared internal Grafana bucket.                    | No       | `grafana-plugin-ci-playwright-reports` |
