# Governance Review: k8sgpt (self-assessment)

What follows is a governance review and assessment for the k8sgpt project, written by the project as a
self-assessment ahead of its CNCF Incubation application, using the TOC
[governance review template](https://github.com/cncf/toc/blob/main/operations/project-reviews/governance-review-template.md).
Evidence is current as of 2026-10-30. Activity figures are approximate counts from GitHub search over the
12 months to 2026-10-30.

Several items below are addressed by open pull requests in k8sgpt-ai; they are marked **(PR pending)** and
must be merged before this review is filed with the TOC.

## Summary and Assessment

**Status:** Needs Work (active maintainers are Alex Jones and Anais Urlichs; expected to reach Mostly Satisfactory once the pending PRs merge and the community and `.project` lists match)

### Executing the Assessment

The review was carried out by reading the project's governance documents in
[k8sgpt-ai/k8sgpt](https://github.com/k8sgpt-ai/k8sgpt), [k8sgpt-ai/community](https://github.com/k8sgpt-ai/community)
and [k8sgpt-ai/.project](https://github.com/k8sgpt-ai/.project), the repository settings file
([.github/settings.yml](https://github.com/k8sgpt-ai/k8sgpt/blob/main/.github/settings.yml)) and CODEOWNERS, and
GitHub activity (PRs, reviews, issues) for k8sgpt and k8sgpt-operator.

### Must-Fix Items

**The following issues have been identified that need to be resolved before the Incubation application:**

1. **Reconcile the maintainer list.** Maintainers have decided that the only active maintainers are Alex Jones
   (@AlexsJones) and Anais Urlichs (@AnaisUrlichs). Everyone else listed as a maintainer moves to emeritus:
   thschue, matthisholleville, bradmccoydev, rakshitgondwal and roberthstrand (listed in
   [MAINTAINERS.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/MAINTAINERS.md)), and arbreezy (Aris Boutselis),
   who is listed in [community/MAINTAINERS.md](https://github.com/k8sgpt-ai/community/blob/main/MAINTAINERS.md) and
   [.project/maintainers.yaml](https://github.com/k8sgpt-ai/.project/blob/main/maintainers.yaml) but not in
   MAINTAINERS.md. thschue stepped down from the community list in September 2024
   ([community#61](https://github.com/k8sgpt-ai/community/pull/61)). MAINTAINERS.md in this repository is updated
   to match that decision (PR pending). The same active list still has to be copied into the community and
   `.project` files. Until those two files change, the lists still disagree.
2. **Maintainer activity.** In the last 12 months, AlexsJones reviewed about 162 PRs in k8sgpt; we found no PR
   reviews from the other listed maintainers. The project needs more than one active reviewer/maintainer, especially
   for k8sgpt-operator, where the median time to first response on new issues is about 35 days (all responses by one maintainer).
3. **Affiliations.** Affiliations in MAINTAINERS.md must be confirmed by each maintainer, and the 30-day update
   policy adopted (PR pending: GOVERNANCE.md).
4. **Sub-team and Security Response Team assignment** must be documented (PRs pending: GOVERNANCE.md, SECURITY.md).
   The Security Response Team is Alex Jones (@AlexsJones) and Anais Urlichs (@AnaisUrlichs).

### Points of Excellence

- Vendor-neutrality is explicit in [GOVERNANCE.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/GOVERNANCE.md#vendor-neutrality),
  including equal treatment of AI backend providers.
- A detailed, Kubernetes-derived contributor ladder exists in
  [community-membership.md](https://github.com/k8sgpt-ai/community/blob/main/community-membership.md).
- Contribution acceptance is enforced by configuration: DCO is a required check and code-owner review is required
  on `main` ([.github/settings.yml](https://github.com/k8sgpt-ai/k8sgpt/blob/main/.github/settings.yml)).

### Areas for Improvement

**Over the next year, the project should work on the following issues to improve its governance, these are considered non-blocking:**

- Grow the reviewer/approver bench in k8sgpt-operator and add a second responder for its issues.
- Settle contributor agreements: operator PRs run both DCO and EasyCLA
  ([k8sgpt-operator#861](https://github.com/k8sgpt-ai/k8sgpt-operator/issues/861)), while `.project` records `has_cla: false`.
- Announce the monthly public community meeting and list it on the CNCF calendar. The meeting is agreed
  (open to everyone, with an agenda and a notes link); the schedule is still to be confirmed. The date, time,
  agenda and notes link will be announced in `#k8sgpt` on CNCF Slack and on the issue tracker.
- Align the subproject list in GOVERNANCE.md with [subprojects.md](https://github.com/k8sgpt-ai/community/blob/main/subprojects.md)
  (GOVERNANCE.md lists docs, charts and website; subprojects.md lists Documentation/Website and Demo and has no charts entry).
- Record governance changes in a changelog section so governance evolution is easy to follow.

---

## Review

**The following review primarily consists of an audit on the project's self-assessment in their Incubation application.**

The project's self-assessment is available at: _(link to the cncf/toc incubation application issue, once filed)_.

### Governance Summary

k8sgpt is governed by its maintainers under [GOVERNANCE.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/GOVERNANCE.md):
consensus-seeking with a 48-hour review window, falling back to a simple-majority vote with a 50% quorum; ties are broken
by the project lead (AlexsJones). Adding or removing maintainers, governance changes, direction changes, deprecations and
licensing changes require a formal vote; governance amendments require a 2/3 supermajority. Subproject roles come from the
org-wide contributor ladder in k8sgpt-ai/community.

### Governance Evolution

**Incubating:** Suggested | **Graduated:** Suggested

- 2023-03/04: Code of Conduct, CODEOWNERS and approver groups added; contributor ladder and subprojects documented in the community repo.
- 2023-12-19: accepted into the CNCF Sandbox ([cncf/sandbox#148](https://github.com/cncf/sandbox/issues/148)).
- 2024-09: maintainer changes in the community repo (thschue stepped down, matthisholleville and arbreezy added;
  [community#61](https://github.com/k8sgpt-ai/community/pull/61), [#62](https://github.com/k8sgpt-ai/community/pull/62), [#63](https://github.com/k8sgpt-ai/community/pull/63)).
- 2026-04-24: GOVERNANCE.md, MAINTAINERS.md, ROADMAP.md and RELEASE.md added ([#1642](https://github.com/k8sgpt-ai/k8sgpt/pull/1642)).
- 2026-07/08: `.project` metadata scaffolded by CNCF automation and maintainers aligned to community/MAINTAINERS.md.
- 2026-10: roadmap restructured ([#1882](https://github.com/k8sgpt-ai/k8sgpt/pull/1882)); governance PRs for sub-teams, affiliations and emeritus (pending).

Assessment: governance is being iterated on, but most formal documents are recent (April 2026), and the history shows the
maintainer list drifting between repositories. Partially met.

### Discoverability

**Incubating:** Suggested | **Graduated:** Required

GOVERNANCE.md, MAINTAINERS.md, CONTRIBUTING.md, CODE_OF_CONDUCT.md and SECURITY.md are at the root of k8sgpt-ai/k8sgpt.
The contributor ladder and subproject list live in k8sgpt-ai/community and are linked from GOVERNANCE.md (PR pending).
Met once the pending PR merges.

### Accuracy and Clarity

**Governance is up to date with actual project activities, including any meetings, elections, leadership, or approval processes.**
**Incubating:** Suggested | **Graduated:** Required

Not met today: community/MAINTAINERS.md and .project/maintainers.yaml still list a different active set than the
decision recorded here (Alex Jones and Anais Urlichs only). Community links point to an old Slack invite (PR pending
to move to `#k8sgpt` on CNCF Slack). A monthly public community meeting is proposed; the schedule is still to be confirmed.

**Governance clearly documents vendor-neutrality of project direction.**
**Incubating:** Suggested | **Graduated:** Required

Met: [GOVERNANCE.md § Vendor Neutrality](https://github.com/k8sgpt-ai/k8sgpt/blob/main/GOVERNANCE.md#vendor-neutrality).

### Decisions and Role Assignments

**Document how the project makes decisions on leadership roles, contribution acceptance, requests to the CNCF, and changes to governance or project goals.**
**Incubating:** Suggested | **Graduated:** Required

Mostly met: leadership, governance and direction decisions are covered by GOVERNANCE.md; contribution acceptance by
CONTRIBUTING.md, CODEOWNERS and branch protection. Requests to the CNCF (e.g. Service Desk tickets, budget) are not
explicitly covered. **TODO (maintainers):** state who may file CNCF requests.

**Document how role, function-based members, or sub-teams are assigned, onboarded, and removed for specific teams (example: Security Response Committee).**
**Incubating:** Suggested | **Graduated:** Required

Met once the pending GOVERNANCE.md PR merges (§ Sub-teams and Role Assignments, § Security Response Team).
The Security Response Team is Alex Jones (@AlexsJones) and Anais Urlichs (@AnaisUrlichs), named in SECURITY.md
(PR pending).

### Maintainers and Maintainer Lifecycle

**Document a complete maintainer lifecycle process (including roles, onboarding, offboarding, and emeritus status).**
**Incubating:** Suggested | **Graduated:** Required

Met once the pending GOVERNANCE.md PR merges: becoming a maintainer, stepping down, inactivity, emeritus process and
revocation are documented.

**Demonstrate usage of the maintainer lifecycle with outcomes, either through the addition or replacement of maintainers as project events have required.**
**Incubating:** Suggested | **Graduated:** Required

Partially met: September 2024 additions and a step-down in the community repo
([#61](https://github.com/k8sgpt-ai/community/pull/61)–[#63](https://github.com/k8sgpt-ai/community/pull/63)). The emeritus list in
MAINTAINERS.md was empty and was not updated for that step-down. The current decision moves every non-active
maintainer to emeritus in MAINTAINERS.md (PR pending). Periods of service are recorded only where a public
step-down already exists (thschue, 2024-09); other periods are not recorded here.

**Document complete list of current maintainers, including names, contact information, domain of responsibility, and affiliation.**
**Incubating:** Required | **Graduated:** Required

Partially met. Active maintainers are Alex Jones (@AlexsJones) and Anais Urlichs (@AnaisUrlichs). MAINTAINERS.md
records handles, affiliations, domains and contacts for them (PR pending) and moves everyone else to emeritus.
Affiliations are the values already in that file (AWS and Independent) and still need confirmation by each person.
community/MAINTAINERS.md and .project/maintainers.yaml are not updated in this repository (Must-Fix 1).

**A number of active maintainers which is appropriate to the size and scope of the project.**
**Incubating:** Required | **Graduated:** Required

Not met today. k8sgpt had 191 merged PRs in 12 months (119 of them from 42 human authors, the rest from bots) and k8sgpt-operator about 65 PRs from 12 authors,
but review and issue response are concentrated in one maintainer (Must-Fix 2).

**Project maintainers from at least 2 organizations that demonstrates survivability.**
**Incubating:** N/A | **Graduated:** Required

N/A for Incubation. **TODO (maintainers):** confirm affiliations.

### Ownership

**Code and Doc ownership in Github and elsewhere matches documented governance roles.**
**Incubating:** Required | **Graduated:** Required

Unverified. CODEOWNERS assigns `*` to `@k8sgpt-ai/maintainers`, `@k8sgpt-ai/k8sgpt-maintainers` and `@k8sgpt-ai/k8sgpt-approvers`;
settings.yml grants `admin`, `maintain`, `push` and `push` to maintainers, k8sgpt-maintainers, k8sgpt-approvers and contributors.
Team membership is not public. **TODO (maintainers):** confirm the team membership matches MAINTAINERS.md after reconciliation.

### Code of Conduct

**Document adoption and adherence to the CNCF Code of Conduct or the project's CoC which is based off the CNCF CoC and not in conflict with it.**
**Incubating:** Required | **Graduated:** Required

Met: GOVERNANCE.md adopts the CNCF Code of Conduct; CODE_OF_CONDUCT.md is Contributor Covenant 2.0, with a CNCF CoC preamble
and escalation path to the CNCF Code of Conduct Committee (PR pending).

**CNCF Code of Conduct is cross-linked from other governance documents.**
**Incubating:** Required | **Graduated:** Required

Met once the pending PR merges (GOVERNANCE.md, CODE_OF_CONDUCT.md and README).

### Subprojects

**All subprojects, if any, are listed.**
**Incubating:** Required | **Graduated:** Required

| Area | Ownership and Operation | Communications | Project Alignment | Notes |
|:-----|:-----------------------:|:---------------:|:------------------|:---|
| k8sgpt (CLI/server) | Complete | Complete | Complete | Main project |
| k8sgpt-operator | Partial | Partial | Complete | Maintained by one person per subprojects.md; slow issue response |
| docs / website | Partial | Partial | Complete | Listed in both documents |
| charts | Partial | Partial | Complete | In GOVERNANCE.md, missing from subprojects.md |
| community | Complete | Complete | Complete | |
| Demo (killercoda) | Partial | Other | Complete | In subprojects.md, missing from GOVERNANCE.md |

**If the project has subprojects: subproject leadership, contribution, maturity status documented, including add/remove process.**
**Incubating:** Suggested | **Graduated:** Required

Partially met: leadership groups are in subprojects.md; maturity status and the add/remove process are not documented.

### Contributors and Community

**Contributor ladder with multiple roles for contributors.**
**Incubating:** Suggested | **Graduated:** Suggested

Met: [community-membership.md](https://github.com/k8sgpt-ai/community/blob/main/community-membership.md) defines five roles.

**Clearly defined and discoverable process to submit issues or changes.**
**Incubating:** Required | **Graduated:** Required

Met: [CONTRIBUTING.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/CONTRIBUTING.md), issue templates in `.github/ISSUE_TEMPLATE`, PR template, DCO guidance.

**Project must have, and document, at least one public communications channel for users and/or contributors.**
**Incubating:** Required | **Graduated:** Required

Met: `#k8sgpt` on CNCF Slack (CONTRIBUTING.md), GitHub issues and discussions.

**List and document all project communication channels, including subprojects (mail list/slack/etc.). List any non-public communications channels and what their special purpose is.**
**Incubating:** Required | **Graduated:** Required

Partially met: public channels are listed in GOVERNANCE.md; links being consolidated on CNCF Slack (PR pending).
**TODO (maintainers):** list any private maintainer or security channels and their purpose.

**Up-to-date public meeting schedulers and/or integration with CNCF calendar.**
**Incubating:** Required | **Graduated:** Required

Partially met. Maintainers will hold a monthly public community meeting, open to everyone, with an agenda and a
notes link. The schedule is still to be confirmed: the date, time, agenda and notes link will be announced in
`#k8sgpt` on CNCF Slack and on the issue tracker, and no meeting URL or time is set yet. k8sgpt has a page on the
CNCF LFX meetings calendar (https://zoom-lfx.platform.linuxfoundation.org/meetings/k8sgpt). One-off public sessions
have been held (for example a CNCF community group session on 2026-08-28).

**Documentation of how to contribute, with increasing detail as the project matures.**
**Incubating:** Required | **Graduated:** Required

Met: CONTRIBUTING.md (workflow, DCO, conventional commits) plus the contributor ladder.

**Demonstrate contributor activity and recruitment.**
**Incubating:** Required | **Graduated:** Required

Met for contributions: in the last 12 months k8sgpt merged 191 PRs (119 from 42 human authors, the rest from bots) and
k8sgpt-operator about 65 PRs from 12 authors; 15 releases (v0.4.27–v0.4.41, 2025-12-18 to 2026-10-21). Recruitment into reviewer/maintainer roles is the gap (see Must-Fix 2).
