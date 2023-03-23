# Contributing

## Requirements
- Golang `1.20`
- An OpenAI API key
  * OpenAI API keys can be obtained from [OpenAI](https://platform.openai.com/account/api-keys)
  * You can set the API key for k8sgpt using `./k8sgpt auth key`
- If you want to build the container image, you need to have a container engine (docker, podman, rancher, etc.) installed

## Building
Building the binary is as simple as running `go build .` in the root of the repository. If you want to build the container image, you can run `docker build -t k8sgpt -f container/Dockerfile .` in the root of the repository.

## Releasing
Releases of k8sgpt are done using [Release Please](https://github.com/googleapis/release-please) and [GoReleaser](https://goreleaser.com/). The workflow looks like this:

* A PR is merged to the `main` branch:
  * Release please is triggered, creates or updates a new release PR
  * This is done with every merge to main, the current release PR is updated every time

* Merging the 'release please' PR to `main`:
  * Release please is triggered, creates a new release and updates the changelog based on the commit messages
  * GoReleaser is triggered, builds the binaries and attaches them to the release
  * Containers are created and pushed to the container registry

> With the next relevant merge, a new release PR will be created and the process starts again

### Manually setting the version
If you want to manually set the version, you can create a PR with an empty commit message that contains the version number in the commit message. For example:

Such a commit can get produced as follows: `git commit --allow-empty -m "chore: release 0.0.3" -m "Release-As: 0.0.3`