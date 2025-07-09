# plugin-ci-workflows examples

> [!WARNING]
>
> Please [read the docs](https://enghub.grafana-ops.net/docs/default/component/grafana-plugins-platform/plugins-ci-github-actions/010-plugins-ci-github-actions) before using any of these workflows in your repository.

This folder contains some examples on how to use the shared workflows (CI and CD) in various scenarios.

The `yaml` files should be put in your repository's `.github/workflows` folder, and customized depending on your needs.

**Each workflow file is just a template/example. Remember to address all the TODOs before starting to use the workflows.**

## [`simple`](./simple/)

A simple setup for a non-provisioned plugin

- CI for each PR and push to main
- Manual deployment to the catalog

## [`provisioned-plugin-manual-deployment`](./provisioned-plugin-manual-deployment/)

An example setup for a provisioned plugin

- CI for each PR and push to main
- Manual deployment to the catalog and Grafana Cloud via Argo workflow + deployment_tools

## [`provisioned-plugin-auto-cd`](./provisioned-plugin-auto-cd/)

An example setup for a provisioned plugin with continuous delivery from the `main` branch

- CI for each PR
- CI + CD for each push to main: deployment to the catalog and Grafana Cloud ("dev", and optionally also "ops") via Argo workflow + deployment_tools
- Manual deployment to the catalog and Grafana Cloud via Argo workflow + deployment_tools (for prod deployment)
