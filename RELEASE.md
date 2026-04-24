# Release Process

This document describes how k8sgpt releases are managed.

## Release Automation

k8sgpt uses automated release tooling:

- **[release-please](https://github.com/googleapis/release-please)** - Tracks changes and manages version bumping
- **[GoReleaser](https://goreleaser.com/)** - Builds binaries, containers, and publishes releases

## Release Workflow

1. **Change Tracking**: Commits are tagged with conventional commit types (`feat:`, `fix:`, `chore:`, etc.)
2. **Release PR**: release-please automatically creates a PR that tracks all changes and bumps the version
3. **Merge to Main**: When the release PR is merged, the release workflow triggers
4. **Build & Publish**: GoReleaser builds binaries for all platforms, creates container images, and publishes the release

### CI Pipeline

The release is managed by [`.github/workflows/release.yaml`](.github/workflows/release.yaml):

1. **Step 1 - release-please**: Analyzes commit history and creates a release PR with changelog
2. **Step 2 - goreleaser**: When a new release is created, builds and publishes:
   - Binaries for Linux (amd64, arm64, 386), macOS (amd64, arm64), Windows (amd64, arm64)
   - Container images (ghcr.io/k8sgpt-ai/k8sgpt)
   - Helm chart updates
   - Homebrew tap updates
   - RPM, DEB, and APK packages

## Release Cadence

k8sgpt aims for **monthly releases**. Minor and patch releases happen as needed for bug fixes and security updates.

## Versioning

k8sgpt follows [Semantic Versioning](https://semver.org/): `MAJOR.MINOR.PATCH`

- **MAJOR**: Incompatible API changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

## Publishing Artifacts

Each release publishes:
- [GitHub Releases](https://github.com/k8sgpt-ai/k8sgpt/releases) with changelog
- Binaries via GitHub Releases
- Container images on [ghcr.io/k8sgpt-ai/k8sgpt](https://ghcr.io/k8sgpt-ai/k8sgpt)
- Helm chart via [k8sgpt-ai/charts](https://github.com/k8sgpt-ai/charts)
- Homebrew formula via [k8sgpt-ai/homebrew-k8sgpt](https://github.com/k8sgpt-ai/homebrew-k8sgpt)
- Krew plugin via [krew registry](https://github.com/k8sgpt-ai/k8sgpt/blob/main/.krew.yaml)
- Package repositories (RPM, DEB, APK)

## Release Configuration

- [release-please-config.json](release-please-config.json) - Configures release-please behavior
- [release-please-manifest.json](release-please-manifest.json) - Tracks current version
- [.goreleaser.yaml](.goreleaser.yaml) - Configures GoReleaser build and publish

## Manual Release

While releases are automated, a manual release can be triggered via:
```bash
# Trigger the release workflow manually
gh workflow run release.yaml -R k8sgpt-ai/k8sgpt
```

## Contributing to Releases

Contributors do not need to manage releases. Just follow conventional commits and the release automation will handle the rest:
- `feat:` - Will bump MINOR version
- `fix:` - Will bump PATCH version
- `chore:`, `docs:`, `refactor:` - Will bump PATCH version
- `BREAKING CHANGE:` in commit body - Will bump MAJOR version
