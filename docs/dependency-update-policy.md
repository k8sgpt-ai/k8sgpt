# Dependency Update Policy

This document describes how k8sgpt keeps its dependencies up to date and how it handles vulnerable
dependencies. It applies to Go modules, GitHub Actions, container base images, and tool versions
referenced in the Makefile and YAML files.

## Tooling

- **[Renovate](https://docs.renovatebot.com/)** is configured in [`renovate.json`](../renovate.json). It opens PRs for
  Go modules (`gomod`), GitHub Actions, Dockerfiles and versions annotated with `# renovate:` comments.
- Non-major updates are grouped (`group:allNonMajor`) and related packages are grouped together
  (Kubernetes, Azure SDK, Prometheus, Go toolchain).
- GitHub Actions are pinned to full commit SHAs (`helpers:pinGitHubActionDigests`), with the version in a comment.
- Renovate commits are signed off (`:gitSignOff`) so that they pass the DCO check.
- `go mod tidy` runs on every Go module update (`postUpdateOptions`).

## Review and merge rules

| Update type | Policy |
|-------------|--------|
| Patch / minor, CI green | Auto-merged by Renovate (`automerge`, `platformAutomerge`) once required checks pass. |
| Major versions | Opened as a PR and reviewed by a maintainer; never auto-merged for `0.x` dependencies. |
| Security fixes (Renovate `[security]` PRs, GitHub advisories) | Prioritised: reviewed within 7 days; Critical/High fixes are released per the timelines in [SECURITY.md](../SECURITY.md). |
| New direct dependencies | Added in a normal PR; reviewers check license compatibility (CNCF [allowed licenses](https://github.com/cncf/foundation/blob/main/policies-guidance/allowed-third-party-license-policy.md)), maintenance status and necessity. |

## Vulnerable dependencies

1. Renovate `[security]` PRs and GitHub Dependabot alerts (TODO maintainers: confirm Dependabot alerts are enabled) are triaged by the
   [Security Response Team](../SECURITY.md#security-response-team).
2. If a vulnerable dependency affects k8sgpt, the fix is merged to `main` and shipped in the next patch release.
3. If the vulnerable code is not reachable, the PR is still merged in the normal flow and the reasoning is noted in the PR.

## License compliance

New or updated dependencies must keep the project compliant with the CNCF third-party license policy.
If a dependency update introduces a non-allowlisted license, the update is held or pinned until it is
resolved (for example [#1850](https://github.com/k8sgpt-ai/k8sgpt/issues/1850)).
