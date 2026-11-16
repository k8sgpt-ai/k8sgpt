# k8sgpt Roadmap

This document describes the direction of k8sgpt and k8sgpt-operator. It is kept in sync with the [k8sgpt-ai project board](https://github.com/orgs/k8sgpt-ai/projects/1), where every roadmap item carries a **Horizon** (Now / Next / Later) and an **Area** field.

_Last updated: 16 November 2026_

## Project Vision

k8sgpt gives every Kubernetes practitioner SRE-grade diagnostics. Our focus is to be the **precise, low-false-positive diagnosis layer** for Kubernetes: humans use it directly through the CLI, it runs continuously in-cluster through the operator, and AI agents call it over MCP.

## Horizons

- **Now**: the next ~3 months
- **Next**: roughly 3 to 6 months
- **Later**: after that, or waiting on design or demand

## k8sgpt (CLI)

### Now
- Keep `main` green and releasable: License Compliance fix (#1850), regular releases
- Security dependency updates that build cleanly (#1791, #1637, #1690, #1837)
- MCP server on mcp-go v1 with structured output (#1788, #1501)
- Analyzer findings as first-class MCP tools and resources (#1878)
- Finish open false-positive fixes (#1581, #1668) and the custom-analyzer filter (#1374)
- Mute known findings via configuration (#372)
- False-positive regression corpus for analyzers (#1879)
- Reliable Azure and private/enterprise OpenAI-compatible endpoints (#1074, #1538)

### Next
- Opt-in, CLI-only action classification behind a feature gate (#1813)
- Generic CRD analyzer (#1528)
- Custom prompt prefix (#842)
- AI backend support tiers: core vs community (#1880)
- Analyzers for new Kubernetes APIs as they mature (DRA, ValidatingAdmissionPolicy, Gateway API)

### Later
- Findings in a standard format (OpenTelemetry logs / CloudEvents) (#1881)
- Bedrock Knowledge Base support (#1514)

## k8sgpt-operator

### Now
- Regular releases in step with the CLI (next: 0.2.30)
- Rebuild images on patched Go and CA bundles (k8sgpt-operator#860)
- Triage open bugs (k8sgpt-operator#769, #752, #311, #717, #816)
- Contributor sign-off on the operator is DCO. [k8sgpt-operator#864](https://github.com/k8sgpt-ai/k8sgpt-operator/pull/864) closed #861 and documents DCO as the only sign-off. Disabling the EasyCLA GitHub App is an organization setting and is not done by that pull request

### Next
- Serve MCP from the operator (k8sgpt-operator#748)
- Keyless cloud auth: IRSA and Entra Workload Identity (k8sgpt-operator#643, #801)
- Parity with CLI configuration: customHeaders, customrest (k8sgpt-operator#517, k8sgpt#1452)
- kubebuilder v4 migration (k8sgpt-operator#421)

### Later
- Custom analysis via the operator (k8sgpt-operator#475)
- Emit findings for agent runtimes and observability stacks (#1881)

## Community and Adoption

- CNCF Incubation preparation is tracked in #1641
- If your organisation uses k8sgpt, please add it to [ADOPTERS.md](ADOPTERS.md)
- New contributors: see [CONTRIBUTING.md](CONTRIBUTING.md); issues labelled `good first issue` are a good place to start

## How the roadmap changes

Propose a change by opening a GitHub issue, or by discussing it in `#k8sgpt` on the [CNCF Slack](https://slack.cncf.io/). Proposals are reviewed at the monthly public community meeting (second Thursday, 16:00–17:00 UK time; see [GOVERNANCE.md](GOVERNANCE.md#community-meeting)).

Maintainers approve changes under [GOVERNANCE.md](GOVERNANCE.md#decision-making). A change of project direction is a significant decision and takes a formal vote.

## How to Contribute

The roadmap is community-driven. To suggest new items:
1. Open a GitHub issue with the `enhancement` label
2. Discuss your idea in `#k8sgpt` on the [CNCF Slack](https://slack.cncf.io/)
3. Submit a PR implementing the feature

## Previous Roadmaps

- 2026-04-24: initial roadmap
- 2026-10-30: restructured into Now / Next / Later, aligned with the project board
- 2026-11-16: added how roadmap items are proposed and approved

## Feedback

We welcome feedback on this roadmap through GitHub Issues or `#k8sgpt` on the [CNCF Slack](https://slack.cncf.io/). GitHub Discussions are not enabled on this repository.
