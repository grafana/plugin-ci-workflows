<!-- THIS FILE IS AUTO-GENERATED. DO NOT EDIT MANUALLY. RUN: "go run genreadme.go" TO RE-GENERATE -->

# provisioned-plugin-auto-cd

> [!WARNING]
>
> Please [read the docs](https://enghub.grafana-ops.net/docs/default/component/grafana-plugins-platform/plugins-ci-github-actions/010-plugins-ci-github-actions) before using any of these workflows in your repository.
The `yaml` files should be put in your repository's `.github/workflows` folder, and customized depending on your needs.

An example setup for a provisioned plugin with continuous delivery from the `main` branch. This is the **recommended** workflow to use for new plugins that want to be automatically installed on Grafana Cloud Instances.

- CI for each PR
- CI + CD for each push to main: deployment to the catalog and Grafana Cloud ("dev", and optionally also "ops") via Argo workflow + deployment_tools
- Manual deployment to the catalog and Grafana Cloud via Argo workflow + deployment_tools (for prod deployment)