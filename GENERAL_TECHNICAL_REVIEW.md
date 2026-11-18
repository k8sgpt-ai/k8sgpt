# General Technical Review - k8sgpt / Incubation

- **Project:** https://github.com/k8sgpt-ai/k8sgpt
- **Project Version:** v0.4.41 (k8sgpt-operator: v0.2.29)
- **Website:** https://k8sgpt.ai
- **Date Updated:** 2026-11-16
- **Template Version:** v1.0
- **Description:** k8sgpt is a tool for scanning Kubernetes clusters, diagnosing, and triaging issues in simple English. It has SRE experience codified into analyzers and helps pull out the most relevant information, enriching it with AI from various LLM providers.

---

## Day 0 - Planning Phase

### Scope

**Describe the roadmap process, how scope is determined for mid to long term features, as well as how the roadmap maps back to current contributions and maintainer ladder?**

The roadmap is documented in [ROADMAP.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/ROADMAP.md). Scope is determined through:
- GitHub Issues and PRs from the community
- Slack discussions in `#k8sgpt` on [CNCF Slack](https://slack.cncf.io)
- The monthly public community meeting (second Thursday, 16:00–17:00 UK time, from 12 Nov 2026; https://meet.google.com/ugf-zqjt-nwu)
- Maintainer retrospectives and planning sessions

GitHub Discussions are not enabled on this repository.

Contributions follow the contributor ladder ([community-membership.md](https://github.com/k8sgpt-ai/community/blob/main/community-membership.md), linked from [GOVERNANCE.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/GOVERNANCE.md)): contributors → members → approvers → maintainers. Roadmap items are prioritized by consensus of the active maintainers (Alex Jones and Anais Urlichs), with feature requests evaluated through GitHub issues; roadmap tracking issues #1878–#1881 are open for contribution.

**Describe the target persona or user(s) for the project?**

- **Junior SREs / DevOps Engineers** who need help diagnosing Kubernetes issues but lack deep domain expertise
- **Experienced SREs** who want to accelerate their troubleshooting workflow
- **Platform Engineering Teams** who want continuous cluster monitoring via the operator
- **Kubernetes Educators** who use k8sgpt to teach cluster health concepts
- **Dev Teams** who need quick insights into cluster issues without waiting for SRE availability

**Explain the primary use case for the project. What additional use cases are supported by the project?**

- **Primary:** Real-time cluster diagnostics — scan a Kubernetes cluster and get AI-powered explanations of issues in plain English
- **Additional:**
  - Continuous monitoring via k8sgpt operator
  - Prometheus metrics from `k8sgpt serve` (`/metrics`) and an optional operator ServiceMonitor (off by default)
  - Custom analyzer framework for extensibility (gRPC)
  - MCP server integration for AI assistant workflows

**Explain which use cases have been identified as unsupported by the project.**

- The k8sgpt CLI and server do not modify or remediate cluster state (read-only by design). The operator's optional auto-remediation (alpha, off by default) is the only component that writes, and only through a deny-by-default policy gate ([AUTO_REMEDIATION.md](https://github.com/k8sgpt-ai/k8sgpt-operator/blob/main/AUTO_REMEDIATION.md)); it is not intended for production changes without independent review
- k8sgpt does not replace comprehensive security scanning tools (focuses on operational issues, not security posture)
- k8sgpt does not provide its own AI model (consumes external AI APIs)
- k8sgpt does not persistently store cluster data (except optional remote caching)

**Describe the intended types of organizations who would benefit from adopting this project.**

- Financial services organizations needing SRE-grade diagnostics
- Cloud-native software manufacturers
- Organizations providing platform engineering services
- Managed Kubernetes providers
- Education and training organizations
- Any organization running Kubernetes at any scale

**Please describe any completed end user research and link to any reports.**

End user feedback is gathered through:
- Slack community discussions (#k8sgpt)
- GitHub Issues and PRs
- Product Hunt user feedback
- KubeCon and CNCF event discussions

No formal end user research reports have been published yet, but the project maintains detailed issue tracking that captures user pain points and feature requests.

### Usability

**How should the target personas interact with your project?**

- **CLI users:** Install via Homebrew, install kubectl plugin via Krew, or download binary. Run `k8sgpt analyze --explain` for instant diagnostics.
- **Operator users:** Deploy via Helm chart for continuous monitoring with Prometheus/Alertmanager integration.
- **Developer users:** Use the custom analyzer framework to write extensible analyzers in any language (Rust, Go, Python, etc.).
- **AI assistant users:** Connect via MCP server to Claude Desktop, ChatGPT, or other MCP-compatible clients.

**Describe the user experience (UX) and user interface (UI) of the project.**

k8sgpt provides:
- **CLI:** Clean terminal output with color-coded results, table formatting, and JSON export. The `--explain` flag provides natural language AI-powered explanations.
- **Operator:** Passive background monitoring with no interactive UI. Results flow through Prometheus metrics and Alertmanager alerts.
- **Web interface:** The [documentation site](https://docs.k8sgpt.ai) provides comprehensive guides and reference materials.

**Describe how this project integrates with other projects in a production environment.**

k8sgpt integrates with:
- **Prometheus:** `k8sgpt serve` exposes `/metrics`. The operator chart can create a ServiceMonitor and a Grafana dashboard; both are off by default (`serviceMonitor.enabled`, `grafanaDashboard.enabled`)
- **Helm:** Deploy via the operator Helm chart
- **GitOps:** The Helm chart can be applied by tools such as Argo CD or Flux
- **MCP:** `k8sgpt serve --mcp` (stdio by default; `--mcp-http` listens on port 8089)
- **Custom analyzers:** external analyzers over the gRPC API in [schema](https://github.com/k8sgpt-ai/schemas/blob/main/protobuf/schema/v1/custom_analyzer.proto) (`pkg/custom/client.go`)

### Design

**Explain the design principles and best practices the project is following.**

- **Vendor neutrality:** Support for 13+ AI backends (OpenAI, Azure, Cohere, Ollama, Amazon Bedrock, Google Gemini, etc.)
- **Read-only analysis:** analyzers never modify cluster state; writes are limited to the operator's opt-in, policy-gated auto-remediation
- **Anonymization (opt-in):** with `--anonymize`, object names and labels are masked before sending to AI backends (events are not yet covered)
- **Extensibility:** Plugin-style analyzer framework for custom analyzers
- **Convention over configuration:** Sensible defaults, optional overrides
- **Modular architecture:** Analyzers are independent and composable

**Outline or link to the project's architecture requirements? Describe how they differ for Proof of Concept, Development, Test and Production environments, as applicable.**

k8sgpt has a single deployment model that works across all environments:
- **PoC/Dev:** Install CLI, point at any kubeconfig, run analysis
- **Test:** Same as Dev, with custom analyzers and filters
- **Production:** Deploy operator via Helm, configure continuous monitoring, integrate with monitoring/alerting tools

The architecture does not change between environments — only the configuration and deployment method differs.

**Define any specific service dependencies the project relies on in the cluster.**

k8sgpt has no in-cluster service dependencies. It reads directly from the Kubernetes API server. Optional integrations (a Prometheus ServiceMonitor, a remote cache, custom gRPC analyzers, an AI backend) are services the user configures.

**Describe how the project implements Identity and Access Management.**

- Uses the user's existing kubeconfig for cluster authentication (in-cluster or local)
- No additional IAM roles or service accounts required
- Operator runs as a standard Kubernetes deployment with its own ServiceAccount and RBAC; it reads cluster resources, manages its own CRDs (`K8sGPT`, `Result`, `Mutation`) and the k8sgpt server Deployment, and only patches workloads when auto-remediation is enabled
- AI backend API keys are managed by the user and stored locally

**Describe how the project has addressed sovereignty.**

- Local AI options (Ollama, LocalAI) keep data within the user's network
- Users can configure any AI backend including self-hosted models
- The opt-in anonymization feature masks object names and labels before sending to external AI providers
- Remote caching (S3, Azure Blob, GCS, Interplex) can be configured with user-controlled storage

**Describe any compliance requirements addressed by the project.**

- Apache 2.0 license compliance
- CNCF Code of Conduct
- OpenSSF Best Practices badge
- Developer Certificate of Origin (DCO) is the required status check on `main` in both repositories. [k8sgpt-operator#864](https://github.com/k8sgpt-ai/k8sgpt-operator/pull/864) documents DCO as the operator's only contributor sign-off. Disabling the EasyCLA GitHub App is an organization setting outside the repository; that PR does not turn the app off
- Dependency license scanning: a `License Compliance` status check runs on pull requests. There is no workflow file for it in this repo. On 16 Nov 2026 the check's details URL was on app.fossa.com, so FOSSA is the integration that posts it. The FOSSA README badge was removed in [#1644](https://github.com/k8sgpt-ai/k8sgpt/pull/1644), so results are not linked from the README (see [#1850](https://github.com/k8sgpt-ai/k8sgpt/issues/1850), [#1851](https://github.com/k8sgpt-ai/k8sgpt/pull/1851)). CLOMonitor's `license_scanning` check fails (16 Nov 2026)
- GitHub CodeQL code scanning (default setup) analyzes Go and GitHub Actions on `main`

**Describe the project's High Availability requirements.**

k8sgpt is a stateless tool:
- **CLI:** Single-process, no HA requirements
- **Operator:** Deployment replicas come from `controllerManager.replicas` (default 1). The chart passes `--leader-elect`, and the manager supports leader election (`--leader-elect`, default off in the binary, on in the chart; `cmd/main.go`)
- **MCP server:** stdio by default. With `--mcp-http` it is a stateless HTTP server on port 8089 with no built-in TLS or authentication

**Describe the project's resource requirements, including CPU, Network and Memory.**

- **CLI:** a single local process. This repository does not publish a measured CPU or memory figure for it. Analysis time depends on cluster size and AI backend response time
- **Operator chart requests/limits** (`chart/operator/values.yaml`): manager 10m CPU / 64Mi memory requested, 500m / 128Mi limited; kube-rbac-proxy 5m / 64Mi requested, 500m / 128Mi limited
- **Network:** Outbound HTTPS to the AI backend when `--explain` is used (local backends may use HTTP; see the self-assessment), outbound to the Kubernetes API
- **Storage:** local config file, optional remote cache (user-configured)

**Describe the project's storage requirements, including its use of ephemeral and/or persistent storage.**

- **Ephemeral:** No persistent storage required. Config stored locally.
- **Optional:** Remote caching (S3, Azure Blob, GCS, Interplex) for analysis results
- **Operator:** The chart does not create a PersistentVolume. It does create `Result` objects, and `Mutation` objects when auto-remediation is enabled

**Please outline the project's API Design:**

- **Kubernetes API:** Read-only access to standard Kubernetes resources (Pods, Deployments, Services, etc.) for analysis; the operator additionally defines the `K8sGPT`, `Result` and `Mutation` CRDs
- **AI Backend API:** Standard REST/HTTP API following OpenAI chat completion format
- **Custom Analyzer API:** gRPC-based protocol defined in [schema](https://github.com/k8sgpt-ai/schemas/blob/main/protobuf/schema/v1/custom_analyzer.proto)
- **MCP Server:** Model Context Protocol for AI assistant integration
- **Release API:** Automated via release-please and GoReleaser

**Describe the project's release processes, including major, minor and patch releases.**

Documented in [RELEASE.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/RELEASE.md):
- Automated via release-please (tracks conventional commits) and GoReleaser
- Regular release cadence: 15 releases in the 12 months to 16 Nov 2026 (v0.4.27 → v0.4.41, median 17 days apart)
- Release archives carry SPDX SBOMs. Keyless cosign signing (Sigstore, GitHub OIDC) of the release checksum file has been in the release workflow since 3 Nov 2026, with verification steps in [RELEASE.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/RELEASE.md); no release has been cut since, so v0.4.41 and earlier are unsigned
- Semantic versioning (MAJOR.MINOR.PATCH)
- Binaries, container images, Helm charts, Homebrew packages, RPM/DEB/APK packages

### Installation

**Describe how the project is installed and initialized, e.g. a minimal install with a few lines of code or does it require more complex integration and configuration?**

- **CLI:** `brew install k8sgpt` or download binary — requires only an AI API key (`k8sgpt auth add`)
- **Operator:** `helm install k8sgpt-operator k8sgpt-ai/k8sgpt` — requires AI backend configuration
- **MCP:** `k8sgpt serve --mcp` uses stdio by default. `--mcp-http` listens on port 8089 on all interfaces, with no built-in TLS or authentication

**How does an adopter test and validate the installation?**

Run `k8sgpt analyze --explain` to verify the tool connects to the cluster and produces analysis results. The operator can be validated by checking Pod status and Prometheus metrics.

### Security

**Please provide a link to the project's cloud native [security self assessment](https://security-assessments.openssf.org/self-assessment/getting-started-self-assessment/).**

[SECURITY_SELF_ASSESSMENT.md](SECURITY_SELF_ASSESSMENT.md) in this repository. The cncf/toc incubation copy is intended to live at `projects/k8sgpt/security-assessment/self-assessment.md`; that path is not in this repository.

**Please review the Cloud Native Security Tenets from TAG Security.**

k8sgpt satisfies these cloud native security tenets:
- **Secure by default:** analysis is read-only and auto-remediation is off by default; anonymization is available but opt-in (`--anonymize`)
- **Least privilege:** the CLI/server only read cluster state; the operator writes only its own resources unless auto-remediation is enabled
- **Defense in depth:** Multiple layers of security (anonymization, TLS, RBAC, DCO, OpenSSF badge)
- **Transparency:** All code is open source, security policy is documented

**Describe how each of the cloud native principles apply to your project.**

- **Secure by default:** read-only analysis by default; sensitive names and labels can be masked before AI queries with `--anonymize`
- **Least privilege:** Read-only Kubernetes API access
- **Zero trust:** Each AI backend call is authenticated independently
- **Defense in depth:** Multiple security layers (anonymization, TLS, RBAC, DCO, OpenSSF)

**How do you recommend users alter security defaults in order to "loosen" the security of the project?**

Anonymization is off unless `--anonymize` is passed, so by default finding text (including object names) is sent to the configured AI backend when `--explain` is used. Users can loosen the operator by enabling auto-remediation and widening its resource allowlist; this is alpha and should not be used for production changes without independent review.

**Security Hygiene**

- **Frameworks/practices:** Go modules with strict versioning, Renovate for automated dependency updates, golangci-lint, DCO enforcement, CODEOWNERS for PR review
- **Security risk evaluation:** All analyzers are reviewed for data exposure. Known risks (event message masking) are tracked in GitHub issues.

**Cloud Native Threat Modeling**

- **Least privilege:** analyzers only read the Kubernetes API. The operator writes its own resources, and with auto-remediation enabled it can patch an allow-listed container image.
- **Certificate rotation:** Not applicable — k8sgpt does not manage certificates. Uses standard Kubernetes API TLS.
- **Secure software supply chain:** SBOM generated via Syft in release workflow. Renovate for automated dependency updates. OpenSSF Best Practices badge (passing). Keyless cosign signing of the release checksum file via Sigstore and GitHub OIDC ([release.yaml](https://github.com/k8sgpt-ai/k8sgpt/blob/main/.github/workflows/release.yaml)). Workflow token permissions default to read-only; [SECURITY-INSIGHTS.yml](https://github.com/k8sgpt-ai/k8sgpt/blob/main/SECURITY-INSIGHTS.yml), a [dependency-update policy](https://github.com/k8sgpt-ai/k8sgpt/blob/main/docs/dependency-update-policy.md) and an [OpenSSF Scorecard workflow](https://github.com/k8sgpt-ai/k8sgpt/blob/main/.github/workflows/scorecard.yaml) are in place (Scorecard 7.2/10 on 16 Nov 2026: https://scorecard.dev/viewer/?uri=github.com/k8sgpt-ai/k8sgpt).

## Day 1 - Installation and Deployment Phase

### Project Installation and Configuration

**Describe what project installation and configuration look like.**

- **CLI:** `brew install k8sgpt` → `k8sgpt auth add --backend openai` → `k8sgpt analyze --explain`
- **Operator:** Helm chart with configurable AI backends, analyzers, and integrations
- **Custom analyzers:** gRPC service (URL and port in config) plus the k8sgpt config

### Project Enablement and Rollback

**How can this project be enabled or disabled in a live cluster?**

- **CLI:** No enable/disable needed — run when needed
- **Operator:** `helm uninstall`, or scale the controller with `--set controllerManager.replicas=0`. The chart has no top-level `enabled` value

**Describe how enabling the project changes any default behavior of the cluster or running workloads.**

The CLI and server are read-only and have no impact on cluster behavior or running workloads. The operator adds its CRDs and a k8sgpt server Deployment; with auto-remediation enabled it can change workload images through normal rollouts.

**Describe how the project tests enablement and disablement.**

k8sgpt CI runs `go test ./...`, which includes CLI and server unit tests. Operator CI runs `make test` (envtest). A Kind cluster is created only when the operator Helm chart changes. There is no Minikube job.

**How does the project clean up any resources created, including CRDs?**

The CLI creates no cluster resources. Helm uninstall removes the release's chart resources. CRDs installed from `chart/operator/crds/` (`K8sGPT`, `Result`, `Mutation`) are not deleted by Helm, and `Result` / `Mutation` objects remain until those CRDs are removed.

### Rollout, Upgrade and Rollback Planning

**How does the project intend to provide and maintain compatibility with infrastructure and orchestration management tools like Kubernetes and with what frequency?**

- Uses client-go for Kubernetes API interaction
- k8sgpt CI runs `go test ./...`. It does not matrix Kubernetes versions. The operator's `make test` uses envtest pinned to Kubernetes 1.26.0 (`ENVTEST_K8S_VERSION`). Operator CI creates a Kind cluster only when the Helm chart changes. `make e2e-remediation` is a Kind test and is not a CI job. Neither workflow runs Minikube
- Frequent releases (15 in the 12 months to 16 Nov 2026) ship client-go updates with the rest of the tree

**Describe how the project handles rollback procedures.**

- **CLI:** Rollback via `brew upgrade k8sgpt@previous` or downloading previous binary
- **Operator:** `helm rollback` to previous release
- **Container images:** Rollback via image tag

**How can a rollout or rollback fail? Describe any impact to already running workloads.**

No impact on running workloads for the CLI/server, which are read-only. For the operator, a failed rollout or rollback affects only k8sgpt's own Deployment and CRDs, unless auto-remediation is enabled, in which case remediation rollouts follow normal Kubernetes rollout semantics.

**Describe any specific metrics that should inform a rollback.**

- Operator crash loops
- Analysis result quality degradation
- AI backend API errors

**Explain how upgrades and rollbacks were tested.**

k8sgpt CI runs `go test ./...` on every PR and on `main`. The operator runs `make test` (envtest, Kubernetes 1.26.0) on pull requests, and a Kind chart install when the chart changes. There is no Minikube job and no matrix of current plus two previous Kubernetes versions in either workflow.

**Explain how the project informs users of deprecations and removals of features and APIs.**

- Changelog in GitHub Releases
- Deprecation warnings in CLI output
- Documentation updates on docs.k8sgpt.ai

**Explain how the project permits utilization of alpha and beta capabilities as part of a rollout.**

k8sgpt uses only stable Kubernetes API versions. Custom analyzers can leverage alpha APIs if users configure them to do so.

## Day 2 - Day-to-Day Operations Phase

### Scalability/Reliability

**Describe how the project increases the size or count of existing API objects.**

The CLI/server create no API objects. The operator creates `Result` objects for findings (and `Mutation` proposals when auto-remediation is enabled), so object count grows with the number of open findings.

**Describe how the project defines Service Level Objectives (SLOs) and Service Level Indicators (SLIs).**

- **SLO:** Analysis completion within expected timeframes (varies by cluster size and AI backend)
- **SLI:** Percentage of successful analysis runs, error rates, AI backend response times

**Describe any operations that will increase in time covered by existing SLIs/SLOs.**

Larger clusters and more analyzers increase analysis time proportionally. The operator can be configured to scan namespaces independently.

**Describe the increase in resource usage in any components as a result of enabling this project.**

- **CLI:** one local process; no measured figure is published in this repository
- **Operator:** the chart requests 64Mi for the manager and 64Mi for kube-rbac-proxy (see resource requirements above)
- **Network:** Outbound HTTPS to the AI backend per explained analysis, plus Kubernetes API traffic

**Describe which conditions enabling / using this project would result in resource exhaustion.**

The CLI and server do not create cluster objects. The operator creates `Result` objects for findings and `Mutation` objects when auto-remediation is enabled, so object count grows with open findings. The chart does not request a PersistentVolume.

**Describe the load testing that has been performed on the project and the results.**

This repository does not contain a load-test report or a measured cluster-size limit. Operator scan interval is configurable on the `K8sGPT` resource.

**Describe the recommended limits of users, requests, system resources, etc.**

The repository does not publish a hard limit on users, requests or cluster size. The CLI runs against whatever the kubeconfig can read. The operator is the in-cluster mode for scheduled scans.

**Describe which resilience pattern the project uses.**

- **CLI:** Stateless, no circuit breaker needed
- **Operator:** Standard Kubernetes deployment with health checks
- **AI backends:** Multiple backend support provides natural fallback

### Observability Requirements

**Describe the signals the project is using or producing, including logs, metrics, profiles and traces.**

- **Logs:** Standard Kubernetes operator logging (info, warn, error levels)
- **Metrics:** Prometheus-compatible metrics for operator mode (analysis duration, error rates)
- **Traces:** Not implemented in this repository
- **Formats:** JSON logs, Prometheus exposition format

**Describe how the project captures audit logging.**

The CLI and server analyzers do not write cluster objects. The operator writes its own CRDs and, when auto-remediation is enabled, can patch a workload image. k8sgpt does not write its own audit log. Kubernetes API server audit logs record the API calls those credentials make.

**Describe any dashboards the project uses or implements.**

k8sgpt does not ship a dashboard. The server exposes Prometheus metrics, and the operator chart can install a Grafana dashboard ConfigMap (`grafanaDashboard.enabled`, off by default). This repository does not send findings to Slack, Discord or Teams.

**Describe how the project surfaces project resource requirements for adopters to monitor cloud and infrastructure costs.**

The operator chart publishes CPU and memory requests and limits (64Mi requested for the manager and for kube-rbac-proxy). The chart does not create a PersistentVolume. Adopters monitor cost with ordinary Kubernetes resource metrics.

**Which parameters is the project covering to ensure the health of the application/service and its workloads?**

k8sgpt's analyzers cover:
- Pod health (crash loops, OOM, failed states)
- Deployment status (replicas, rollout status)
- Service connectivity
- Resource limits and requests
- Storage issues
- Network policy violations
- Security misconfigurations

**How can an operator determine if the project is in use by workloads?**

k8sgpt is a monitoring tool, not a workload. Operators can check:
- k8sgpt Pod status (operator mode)
- Analysis result metrics (Prometheus)
- `Result` objects created by the operator

**How can someone using this project know that it is working for their instance?**

Run `k8sgpt analyze --explain` and verify analysis results are returned. For the operator, check the logs and Prometheus metrics.

**Describe the SLOs (Service Level Objectives) for this project.**

- **Availability:** k8sgpt CLI is always available (local tool). Operator availability depends on Kubernetes deployment.
- **Performance:** Analysis completes within expected timeframes based on cluster size and AI backend response time.
- **Accuracy:** Analyzer results are based on well-established Kubernetes failure patterns.

**What are the SLIs (Service Level Indicators) an operator can use to determine the health of the service?**

- Operator Pod health (running, not crash-looping)
- Analysis completion rate
- Error rate from AI backends
- Time between analysis scans (operator mode)

### Dependencies

**Describe the specific running services the project depends on in the cluster.**

k8sgpt has no in-cluster dependencies. It reads directly from the Kubernetes API server.

**Describe the project's dependency lifecycle policy.**

Documented in [docs/dependency-update-policy.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/docs/dependency-update-policy.md):
- Renovate opens PRs for Go modules, GitHub Actions, Dockerfiles and `# renovate:` pins. Non-major updates are grouped. Patch and minor updates auto-merge when CI is green. Major updates, and `0.x` dependencies, are reviewed and are not auto-merged
- Security updates (Renovate `[security]` PRs and GitHub advisories) are reviewed within 7 days. Critical and High fixes follow the timelines in [SECURITY.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/SECURITY.md). <!-- TODO (maintainers): confirm Dependabot alerts are enabled. The dependency-update policy already marks this as unconfirmed. -->
- A reachable vulnerable dependency is merged to `main` and shipped in the next patch release

**How does the project incorporate and consider source composition analysis as part of its development and security hygiene?**

- A `License Compliance` status check on pull requests. There is no FOSSA GitHub Actions workflow in this repository. The check's details URL on 16 Nov 2026 was on app.fossa.com, so FOSSA posts the check. The README badge that linked the scan was removed in [#1644](https://github.com/k8sgpt-ai/k8sgpt/pull/1644) (see [#1850](https://github.com/k8sgpt-ai/k8sgpt/issues/1850) and [#1851](https://github.com/k8sgpt-ai/k8sgpt/pull/1851))
- Renovate for dependency tracking
- SPDX SBOMs from GoReleaser (release archives) and Syft (release container image)
- OpenSSF Best Practices badge
- GitHub CodeQL default setup (`Analyze (go)` and `Analyze (actions)` on `main`; no CodeQL workflow file in the repo)

**Describe how the project implements changes based on source composition analysis (SCA) and the timescale.**

Follows [docs/dependency-update-policy.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/docs/dependency-update-policy.md):
- A dependency update that introduces a license outside the CNCF allow list is held or pinned until it is resolved (for example [#1850](https://github.com/k8sgpt-ai/k8sgpt/issues/1850) and the MPL-2.0 pin in [#1851](https://github.com/k8sgpt-ai/k8sgpt/pull/1851))
- Security PRs are reviewed within 7 days. If the vulnerable code affects k8sgpt, the fix ships in the next patch release. Critical and High timelines are in SECURITY.md

### Troubleshooting

**How does this project recover if a key component or feature becomes unavailable?**

- **AI backend unavailable:** Use `--explain=false` to get raw analysis without AI explanations
- **Multiple backends:** Configure fallback backends
- **Kubernetes API:** Standard client-go retry logic handles temporary API server unavailability

**Describe the known failure modes.**

- AI backend API rate limits or outages (mitigated by local models or caching)
- Large cluster analysis may take longer than expected (mitigated by namespace filtering)
- Event message anonymization is not yet complete (tracked in issue #560)

### Compliance

**What steps does the project take to ensure that all third-party code and components have correct and complete attribution and license notices?**

- Apache-2.0 `LICENSE` at the repository root. There is not a `LICENSE` file in every subdirectory
- Go module metadata in `go.mod`
- Apache copyright headers on many Go files. Not every file has one (for example `pkg/ai/customrest.go` does not)
- New and updated dependencies are checked against the CNCF allowed-third-party-license policy, as described in [docs/dependency-update-policy.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/docs/dependency-update-policy.md). License scanning is the external `License Compliance` check described above, not a workflow file in this repo

**Describe how the project ensures alignment with CNCF recommendations for attribution notices.**

- The root `LICENSE` file and Go module attribution
- Apache copyright headers where source files include them
- The dependency-update policy holds or pins a dependency whose license is outside the CNCF allow list

## Day 2 - Security

### Security Hygiene

**How is the project executing access control?**

- GitHub CODEOWNERS enforces PR review requirements
- Two active maintainers, Alex Jones (@AlexsJones) and Anais Urlichs (@AnaisUrlichs), listed in [MAINTAINERS.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/MAINTAINERS.md) (single source of truth); thschue, matthisholleville, bradmccoydev, rakshitgondwal, roberthstrand and arbreezy are emeritus; the org-level lists in [k8sgpt-ai/community MAINTAINERS.md](https://github.com/k8sgpt-ai/community/blob/main/MAINTAINERS.md) and [k8sgpt-ai/.project maintainers.yaml](https://github.com/k8sgpt-ai/.project/blob/main/maintainers.yaml) match
- Workflow tokens default to read-only. `.github/settings.yml` declares branch protection on `main`: one approving review, code-owner review, resolved conversations, linear history, `enforce_admins`, and a required `DCO` check. <!-- TODO (maintainers): confirm these rules are applied on GitHub. The branch-protection API returned HTTP 403 to the token used on 16 Nov 2026, and OpenSSF Scorecard's Branch-Protection check errored while reading classic branch protection. Also confirm org-level 2FA enforcement. -->
- DCO enforcement on all commits
- Branch protection requires PR review and passing CI
- GitHub organization membership controls write access

### Cloud Native Threat Modeling

**How does the project ensure its security reporting and response team is representative of its community diversity (organizational and individual)?**

The Security Response Team is Alex Jones (@AlexsJones) and Anais Urlichs (@AnaisUrlichs), named in
[SECURITY.md § Security Response Team](https://github.com/k8sgpt-ai/k8sgpt/blob/main/SECURITY.md#security-response-team); affiliations are recorded in [MAINTAINERS.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/MAINTAINERS.md).
With two members the team is small and not yet organizationally diverse. [GOVERNANCE.md § Security Response Team](https://github.com/k8sgpt-ai/k8sgpt/blob/main/GOVERNANCE.md#security-response-team) requires at least two members drawn from the maintainers, so the team grows as maintainers are added through the contributor ladder. <!-- TODO (Alex, Anais): confirm current affiliations. -->

**How does the project invite and rotate security reporting team members?**

Security reports are received through GitHub private vulnerability reporting and contact@k8sgpt.ai (not Slack).
Members are assigned, onboarded, rotated and removed as described in
[GOVERNANCE.md § Security Response Team](https://github.com/k8sgpt-ai/k8sgpt/blob/main/GOVERNANCE.md#security-response-team);
the triage lead rotates per report. The response process is in [SECURITY.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/SECURITY.md).
