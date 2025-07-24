# Plugin CI/CD Workflows

This repository contains reusable GitHub Actions workflows and individual actions for building, testing, and deploying Grafana plugins. It provides a standardized CI/CD pipeline that can be easily integrated into any Grafana plugin repository. This repository mostly contains workflows that are relevant to Grafana Labs **internal** teams working on plugins.

For github actions that are suited for external developers see [plugin-actions](https://github.com/grafana/plugin-actions) repository.

## 🎯 Purpose

The plugin-ci-workflows repository centralizes and standardizes the CI/CD process for Grafana plugins, providing:

- **Consistent Build Process**: Standardized frontend and backend build steps
- **Automated Testing**: Unit tests, E2E tests with Playwright, and documentation testing
- **Security Scanning**: Integrated Trufflehog secret scanning
- **Plugin Packaging**: Automated signing and ZIP packaging for multiple architectures
- **Deployment Pipeline**: Automated publishing to Grafana Plugin Catalog and Grafana Cloud
- **Documentation Publishing**: Automated docs publishing to Grafana website


## 🚀 Quick Start

### 1. Basic CI Setup
Start with [examples of how this workflows can be used](https://github.com/grafana/plugin-ci-workflows/tree/main/examples/base).


### 2. Extra CI helpers
Can be [found here](https://github.com/grafana/plugin-ci-workflows/tree/main/examples/extra)

## 📖 Documentation

- **EngHub Documentation**: [Plugins CI GitHub Actions](https://enghub.grafana-ops.net/docs/default/component/grafana-plugins-platform/plugins-ci-github-actions/010-plugins-ci-github-actions)
- **Examples**: See `examples/` and `./github/workflows` directories for complete workflow examples
- **Action Documentation**: Each action includes detailed input/output documentation

## 🤝 Contributing

This repository is maintained by the Grafana Plugins Platform team (`@grafana/plugins-platform`).

### Development Workflow

1. Make changes to workflows or actions
2. Test changes using the examples in your fork
3. Submit a pull request with your changes
4. Automated release via release-please upon merge

### Release Process

- Uses [release-please](https://github.com/googleapis/release-please) for automated releases
- Follows semantic versioning
- Separate versioning for different components
- Automatic changelog generation


## 🔗 Related Resources

- [Grafana Plugin Development Portal](https://grafana.com/developers/plugin-tools)

## 📄 License

This project follows Grafana's licensing terms. Please refer to the repository's license file for details.
