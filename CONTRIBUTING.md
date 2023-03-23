# Contributing

## Getting Started

If you would like to contribute you to the project, please follow the steps below.
1. Introduce yourself on slack or open an issue to let us know you are interested in contributing.
2. Fork the project and clone it locally.
3. Create a branch and follow [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) guidelines for work undertaken.
4. Pull request your changes back to the upstream repository and follow follow the [pull request template](.github/pull_request_template.md) guidelines.

## Release process with release-please

This project uses [release-please](https://github.com/googleapis/release-please) to automate the release process. The release process is triggered by a GitHub Action that runs on a schedule. The schedule is defined in the [release-please.yml](.github/workflows/release.yml) file.

The release process will create a new release and tag on the repository. It will also create a pull request to update the [CHANGELOG.md](CHANGELOG.md) file. The pull request will need to be merged before the next release is created.



## Requirements

- Golang `1.20`
