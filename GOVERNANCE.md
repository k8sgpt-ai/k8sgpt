# k8sgpt Governance

This document describes how k8sgpt is governed and how decisions are made.

## Principles

k8sgpt operates under the following principles:

- **Open**: k8sgpt is open source. All changes to the project are made publicly.
- **Neutral**: k8sgpt is vendor-neutral. No single organization controls the project.
- **Collaborative**: We welcome contributions from all individuals and organizations.
- **Merit-based**: Influence is earned through sustained, quality contributions.
- **Community-focused**: We serve the cloud native community, not any single vendor.

k8sgpt is a CNCF project and adheres to the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).

## Decision Making

### Consensus-Seeking

The project strives for consensus among maintainers. For most decisions:
1. A proposal is made (via GitHub issue or PR)
2. Maintainers review and discuss for at least 48 hours
3. If consensus is reached, the proposal is implemented
4. If consensus cannot be reached, the proposal is put to a vote

### Voting

When consensus cannot be reached on significant matters:
- Each maintainer receives one vote
- A simple majority (50% + 1) decides the outcome
- Quorum requires at least 50% of active maintainers to participate
- Ties are broken by the project lead (AlexsJones)

### Significant Decisions

The following require a formal vote:
- Addition or removal of maintainers
- Changes to this governance document
- Project direction changes (e.g., major architectural shifts)
- Deprecation or removal of core features
- Licensing changes

### Operational Decisions

Minor operational decisions (bug fixes, documentation, small features) do not require a full vote and can be handled through normal PR review.

## Maintainer Lifecycle

### Becoming a Maintainer

1. Make sustained, quality contributions over a period of at least 3 months
2. Be nominated by an existing maintainer
3. Receive approval from a majority of current maintainers
4. Be added to the [MAINTAINERS.md](MAINTAINERS.md) file and granted write access

### Stepping Down

Maintainers may step down at any time by notifying the remaining maintainers. Stepping-down maintainers are listed in the [Emeritus Maintainers](#emeritus-maintainers) section of MAINTAINERS.md.

### Inactive Maintainers

A maintainer who has not contributed (code, review, or community) for 6 months may be asked to step down. If they do not respond within 30 days, the remaining maintainers may vote to move them to emeritus status.

### Revocation

A maintainer's access may be revoked for violations of the CNCF Code of Conduct. This requires a supermajority (2/3) vote of active maintainers.

## Vendor Neutrality

k8sgpt is committed to vendor neutrality:

- No single vendor may control project direction
- All AI backend providers are supported equally (OpenAI, Azure, Cohere, Ollama, Amazon Bedrock, Google Gemini, etc.)
- Decisions about supported backends are made based on community merit, not vendor influence
- Financial contributions from sponsors do not buy decision-making power
- The project benefits from the CNCF's neutral governance structure

## Subproject Governance

k8sgpt has several subprojects under the [k8sgpt-ai organization](https://github.com/k8sgpt-ai):

- **k8sgpt** (main project) - This repository
- **k8sgpt-operator** - Kubernetes operator for continuous monitoring
- **docs** - Documentation site
- **charts** - Helm charts
- **website** - Project website
- **community** - Community management

Subprojects have their own maintainers but are expected to align with the main project's governance. Changes to a subproject's governance should be coordinated with the main project maintainers.

## Community Engagement

We encourage broad community participation:

- **Slack**: [#k8sgpt](https://join.slack.com/t/k8sgpt/shared_invite/zt-332vhyaxv-bfjJwHZLXWVCB3QaXafEYQ)
- **GitHub Issues**: All feature requests and bug reports
- **GitHub Discussions**: Open for community questions and ideas
- **CNCF Slack**: [#k8sgpt](https://slack.cncf.io/)

## Code of Conduct

All participants in this project are expected to adhere to the [CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md). Reports should be sent to the project maintainers and/or CNCF staff.

## Amendment Process

This governance document may be amended at any time by a supermajority (2/3) vote of active maintainers. Proposed amendments should be discussed publicly before voting.
