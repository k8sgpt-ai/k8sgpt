# Security Policy

This policy covers the repositories in the [k8sgpt-ai](https://github.com/k8sgpt-ai) GitHub organization,
including the k8sgpt CLI/server, [k8sgpt-operator](https://github.com/k8sgpt-ai/k8sgpt-operator) and the
[Helm charts](https://github.com/k8sgpt-ai/charts).

## Supported Versions

k8sgpt does not maintain release branches. Security fixes are made on `main` and shipped in a new
release from the latest release line.

| Component | Supported for security fixes |
|-----------|------------------------------|
| k8sgpt CLI / server | Latest `v0.4.x` release |
| k8sgpt-operator | Latest release |
| Helm charts | Latest chart version |

Older releases do not receive backported fixes. Please upgrade to the latest release.

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues, discussions, pull requests or Slack.**

Report privately through either channel:

1. **GitHub private vulnerability reporting (preferred):** open
   [Report a vulnerability](https://github.com/k8sgpt-ai/k8sgpt/security/advisories/new) in this repository.
   Use this link for issues in the operator and Helm charts too; the team will move the advisory to the
   affected repository if needed.
2. **Email:** contact@k8sgpt.ai. Put `[SECURITY]` in the subject line.

Please include as much of the following as you can:

- The affected component and version (`k8sgpt version`, operator or chart version)
- A description of the issue and its impact
- Steps to reproduce or a proof of concept
- Any known mitigations or workarounds
- Whether and how you would like to be credited

## Security Response Team

Security reports are handled by the k8sgpt Security Response Team.

- **Security response team:** Alex Jones (@AlexsJones) and Anais Urlichs (@AnaisUrlichs)

The team is drawn from the maintainers listed in [MAINTAINERS.md](MAINTAINERS.md) and has at least two members.
How members are assigned, rotated and removed is described in [GOVERNANCE.md](GOVERNANCE.md#security-response-team).

### Roles

| Role | Responsibility |
|------|----------------|
| **Triage lead** (rotates per report) | Acknowledges the report, confirms the issue, assigns severity (CVSS v3.1/v4), and coordinates until the report is closed. |
| **Fix owner** | Develops and reviews the fix in a private fork / GitHub security advisory workspace. |
| **Disclosure coordinator** | Requests the CVE through the GitHub advisory, drafts the advisory and release notes, and coordinates the disclosure date with the reporter. |

One person may hold more than one role for a given report, but each fix must be reviewed by a second team member.

### Escalation

If you have not received an acknowledgement within the time below, email contact@k8sgpt.ai again and
mention a maintainer from [MAINTAINERS.md](MAINTAINERS.md).

## Response Process and Timelines

| Step | Target |
|------|--------|
| Acknowledge the report | Within 3 business days |
| Initial assessment and severity | Within 7 days of acknowledgement |
| Fix for Critical / High severity | Target within 30 days |
| Fix for Medium / Low severity | Target within 90 days, or the next scheduled release |
| Public disclosure | When a fixed release is available, or 90 days after the report, whichever is first, unless agreed otherwise with the reporter |

The team keeps the reporter updated at least every 14 days until the report is resolved.

1. **Triage:** the triage lead confirms the report in the private advisory and assigns severity.
2. **Fix:** the fix owner prepares a patch in the advisory's private fork; a second team member reviews it.
3. **Release:** a patch release is published through the normal [release process](RELEASE.md).
   GoReleaser checksums for that release are signed with [cosign](https://docs.sigstore.dev/cosign/signing/overview/) keyless signing (Sigstore, GitHub OIDC). The signature bundle (`checksums.txt.sigstore.json`) is published with the GitHub release. That signing step is added in the release workflow; releases published before it is on `main` are not signed. The container image is built in a separate job and is not covered by this signature.
4. **Disclosure:** the GitHub Security Advisory is published (with a CVE where applicable) and linked from the release notes.
   Users are notified through GitHub Releases and the `#k8sgpt` channel on [CNCF Slack](https://slack.cncf.io).

## Past Advisories

Published advisories are listed at <https://github.com/k8sgpt-ai/k8sgpt/security/advisories>.

## Code of Conduct

Security reports that concern behaviour rather than software should follow the
[Code of Conduct](CODE_OF_CONDUCT.md) process instead.
