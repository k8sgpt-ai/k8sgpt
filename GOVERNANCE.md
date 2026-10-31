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

## Contributor Ladder

k8sgpt uses the contributor ladder defined in the community repository:
[community-membership.md](https://github.com/k8sgpt-ai/community/blob/main/community-membership.md).
It describes the roles **New Contributor → Member → Approver → Subproject Maintainer → Organization Maintainer**,
their responsibilities, and the requirements for each step. The steps below for becoming a maintainer
are the final rung of that ladder.

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

Maintainers review the activity of everyone listed in [MAINTAINERS.md](MAINTAINERS.md) at least once a year,
and before any CNCF maturity-level application.

### Emeritus Maintainers

A maintainer moves to emeritus status when they step down, or after the inactivity process above. To do so:

1. Open a PR that moves the person from **Current Maintainers** to **Emeritus Maintainers** in
   [MAINTAINERS.md](MAINTAINERS.md), recording the period they served. The same PR (or linked PRs) updates the
   org-level lists that mirror it ([community/MAINTAINERS.md](https://github.com/k8sgpt-ai/community/blob/main/MAINTAINERS.md)
   and [.project/maintainers.yaml](https://github.com/k8sgpt-ai/.project/blob/main/maintainers.yaml)).
2. Remove their membership from the maintainer GitHub teams, CODEOWNERS entries, and any sub-team
   (including the Security Response Team), and rotate any shared credentials they held.
3. Ask CNCF staff to update https://maintainers.cncf.io and the maintainer mailing lists.

Emeritus maintainers have no voting rights or write access. They are thanked publicly and may return
through an expedited nomination: a maintainer nominates them and a simple majority of active maintainers approves,
without the 3-month contribution requirement.

### Maintainer Affiliations

Every maintainer's employer or affiliation (or "Independent") is recorded in [MAINTAINERS.md](MAINTAINERS.md).
Maintainers must update their affiliation **within 30 days** of a change in employment or affiliation, by opening
a PR against MAINTAINERS.md. During the annual activity review, maintainers also confirm that their listed
affiliation is still correct. If an affiliation is found to be out of date, any maintainer may open the PR, and
the correction is noted in the PR description.

### Revocation

A maintainer's access may be revoked for violations of the CNCF Code of Conduct. This requires a supermajority (2/3) vote of active maintainers.

## Vendor Neutrality

k8sgpt is committed to vendor neutrality:

- No single vendor may control project direction
- All AI backend providers are supported equally (OpenAI, Azure, Cohere, Ollama, Amazon Bedrock, Google Gemini, etc.)
- Decisions about supported backends are made based on community merit, not vendor influence
- Financial contributions from sponsors do not buy decision-making power
- The project benefits from the CNCF's neutral governance structure

## Sub-teams and Role Assignments

Some responsibilities are held by named sub-teams rather than by all maintainers.

| Team | Purpose | Membership recorded in |
|------|---------|------------------------|
| Organization maintainers | Project direction, governance, org administration | [MAINTAINERS.md](MAINTAINERS.md); GitHub team `@k8sgpt-ai/maintainers` |
| Subproject maintainers | Direction and merges for one subproject | [subprojects.md](https://github.com/k8sgpt-ai/community/blob/main/subprojects.md); subproject GitHub team (e.g. `@k8sgpt-ai/k8sgpt-maintainers`) |
| Approvers | Review and approve contributions to a subproject | Subproject `-approvers` GitHub team (e.g. `@k8sgpt-ai/k8sgpt-approvers`) |
| Security Response Team | Triage and fix vulnerability reports (see [SECURITY.md](SECURITY.md)) | [SECURITY.md](SECURITY.md#security-response-team) |
| Release managers | Approve release-please PRs and oversee releases (see [RELEASE.md](RELEASE.md)) | Maintainers of the subproject being released |

### Assigning, onboarding and removing sub-team members

- **Assignment:** any maintainer may nominate a person (themselves included) for a sub-team in a public issue or PR.
  Membership is approved by lazy consensus of the organization maintainers after 5 business days, or by a simple
  majority vote if anyone objects.
- **Onboarding:** the PR that records the new member also adds them to the matching GitHub team(s) and
  CODEOWNERS entries. An existing team member walks them through the team's process and private channels.
- **Removal:** members leave a sub-team by stepping down, by moving to emeritus status, or by a simple majority vote of
  organization maintainers. Access is removed in the same PR that updates the membership list.
- **Review:** sub-team membership is reviewed with the annual maintainer activity review.

### Security Response Team

- The Security Response Team has at least two members, drawn from the maintainers.
- Current members are Alex Jones (@AlexsJones) and Anais Urlichs (@AnaisUrlichs).
- Members must have two-factor authentication enabled and be reachable through the private reporting channels
  in [SECURITY.md](SECURITY.md).
- The team rotates the triage lead role per report, follows the response process and timelines in
  [SECURITY.md](SECURITY.md), and reports a summary of handled advisories to the maintainers after disclosure.
- Changes to the team follow the assignment and removal process above, and are reflected in SECURITY.md.

## Subproject Governance

k8sgpt has several subprojects under the [k8sgpt-ai organization](https://github.com/k8sgpt-ai):

- **k8sgpt** (main project) - This repository
- **k8sgpt-operator** - Kubernetes operator for continuous monitoring
- **docs** - Documentation site
- **charts** - Helm charts
- **website** - Project website
- **community** - Community management

Subprojects, their repositories and their maintainer/approver groups are listed in [subprojects.md](https://github.com/k8sgpt-ai/community/blob/main/subprojects.md). Subprojects have their own maintainers but are expected to align with the main project's governance. Changes to a subproject's governance should be coordinated with the main project maintainers.

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
