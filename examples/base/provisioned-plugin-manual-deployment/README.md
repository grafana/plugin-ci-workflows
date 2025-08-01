# provisioned-plugin-manual-deployment

> [!WARNING]
>
> Please [read the docs](https://enghub.grafana-ops.net/docs/default/component/grafana-plugins-platform/plugins-ci-github-actions/010-plugins-ci-github-actions) before using any of these workflows in your repository.

The `yaml` files should be put in your repository's `.github/workflows` folder, and customized depending on your needs.

**Each workflow file is just a template/example. Remember to address all the TODOs before starting to use the workflows.**

<!-- README start -->

An example setup for a provisioned plugin. Use this workflow if you wish to have manual control of version rollouts.

- CI for each PR and push to main
- Manual deployment to the catalog and Grafana Cloud via Argo workflow + deployment_tools
