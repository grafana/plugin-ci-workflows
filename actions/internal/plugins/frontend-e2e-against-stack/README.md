# Run e2e tests from frontend plugins against specific stack

This is a GitHub Action that help the execution of e2e tests on any frontend plugin that is using [Playwright](https://playwright.dev/) against specific selected stack.
You need to define in which region the selected stack belong, the plugin from where are executed the tests and optionally which other plugins and datasources you want to provision when starting a Grafana instance.
Also, you need to have the **playwright** configuration and the test specifications in the plugin that run the tests and the action will do the rest.
This action use the following input parameters to run:

| Name                  | Description                                                                                                        | Default           | Required |
| --------------------- |--------------------------------------------------------------------------------------------------------------------|-------------------|----------|
| `plugin-directory`    | Directory of the plugin, if not in the root of the repository. If provided, package-manager must also be provided. | .                 | No       |
| `package-manager`     | The package manager to use for building the plugin                                                                 |                   | No       |
| `npm-registry-auth`   | Whether to authenticate to the npm registry in Google Artifact Registry                                            | false             | No       |
| `stack_slug`          | Name of the stack where you want to run the tests                                                                  |                   | Yes      |
| `env`                 | Region of the stack where you want to run the tests                                                                |                   | Yes      |
| `other_plugins`       | List of other plugins that you want to enable separated by comma                                                   |                   | No       |
| `datasource_ids`      | List of data sources that you want to enable separated by comma                                                    |                   | No       |
| `upload_report_path ` | Name of the folder where you want to store the test report                                                         | playwright-report | No       |
| `upload_videos_path`  | Name of the folder where you want to store the test videos                                                         | playwright-videos | No       |
| `plugin-secrets`      | A JSON string containing key-value pairs of specific plugin secrets necessary to run the tests.                    |                   | No       |
| `grafana-ini-path`    | Path to a custom grafana.ini file to configure the Grafana instance                                                |                   | No       |

## Example workflows

This is an example of how you could use this action.

```yml
name: Build and Test PR

on:
  pull_request:

jobs:
  e2e-tests:
    permissions:
      contents: write
      id-token: write
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          persist-credentials: false

      - name: Get plugin specific secrets
        id: create-plugin-secrets
        shell: bash
        run: |
          echo 'plugin-json-secrets={"MY_SECRET1": "value_secrete_1", "MY_SECRET2": "value_secret_2"}' >> "$GITHUB_OUTPUT"

      - name: Run e2e cross app tests
        id: e2e-cross-apps-tests
        uses: grafana/plugin-ci-workflows/actions/internal/plugins/frontend-e2e-against-stack@main
        with:
          npm-registry-auth: "true"
          stack_slug: "mygrafanastack"
          env: "dev-central"
          other_plugins: "grafana-plugin1-app,grafana-plugin2-app"
          datasource_ids: "grafanacloud-mygrafanastack-prom,grafanacloud-mygrafanastack-logs"
          upload_report_path: "playwright-cross-apps-report"
          upload_videos_path: "playwright-cross-apps-videos"
          plugin-secrets: ${{ steps.create-plugin-secrets.outputs.plugin-json-secrets }}
          grafana-ini-path: "provisioning/custom-grafana.ini"  # Optional
```
