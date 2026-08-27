# Upload Report Artifacts Action

Uploads Playwright test reports and summaries as GitHub artifacts. Used together with the `publish-report` action to upload reports to GCS and post PR comment links.

## Inputs

| Input Name                  | Description                                                                                                           | Required | Default                               |
| --------------------------- | --------------------------------------------------------------------------------------------------------------------- | -------- | ------------------------------------- |
| `test-outcome`              | Outcome of the test step. For example `${{ steps.run-tests.outcome }}`.                                               | Yes      | N/A                                   |
| `grafana-image`             | Grafana image used in the test.                                                                                       | Yes      | `${{ matrix.GRAFANA_IMAGE.NAME }}`    |
| `grafana-version`           | Grafana version used in the test.                                                                                     | Yes      | `${{ matrix.GRAFANA_IMAGE.VERSION }}` |
| `artifact-prefix`           | Prefix for the artifact name.                                                                                         | No       | `gf-playwright-report-`               |
| `upload-report`             | Whether to upload the report directory regardless of outcome.                                                         | Yes      | `true`                                |
| `upload-successful-reports` | Whether to include the full report in the artifact when tests succeeded.                                              | No       | `false`                               |
| `report-dir`                | Directory in which the report is stored.                                                                              | Yes      | `playwright-report`                   |
| `plugin-name`               | Name of the plugin being tested. Useful in mono-repos when multiple plugins generate multiple reports.                | No       | N/A                                   |

## Outputs

| Output Name | Description                   |
| ----------- | ----------------------------- |
| `artifact`  | Name of the uploaded artifact. |
