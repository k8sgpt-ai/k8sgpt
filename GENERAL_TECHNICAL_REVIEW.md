# General Technical Review - k8sgpt / Incubation

- **Project:** https://github.com/k8sgpt-ai/k8sgpt
- **Project Version:** v0.4.32
- **Website:** https://k8sgpt.ai
- **Date Updated:** 2026-04-24
- **Template Version:** v1.0
- **Description:** k8sgpt is a tool for scanning Kubernetes clusters, diagnosing, and triaging issues in simple English. It has SRE experience codified into analyzers and helps pull out the most relevant information, enriching it with AI from various LLM providers.

---

## Day 0 - Planning Phase

### Scope

**Describe the roadmap process, how scope is determined for mid to long term features, as well as how the roadmap maps back to current contributions and maintainer ladder?**

The roadmap is documented in [ROADMAP.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/ROADMAP.md). Scope is determined through:
- GitHub Issues and PRs from the community
- Slack discussions in [#k8sgpt](https://join.slack.com/t/k8sgpt/shared_invite/zt-332vhyaxv-bfjJwHZLXWVCB3QaXafEYQ)
- GitHub Discussions
- Maintainer retrospectives and planning sessions

Contributions follow a clear ladder documented in [GOVERNANCE.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/GOVERNANCE.md): contributors → approvers → maintainers. Roadmap items are prioritized by maintainer consensus, with feature requests evaluated through GitHub issues.

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
  - Slack/Teams integration for alerting
  - Prometheus/Alertmanager integration for monitoring
  - Custom analyzer framework for extensibility
  - MCP server integration for AI assistant workflows

**Explain which use cases have been identified as unsupported by the project.**

- k8sgpt does not modify or remediate cluster state (read-only by design)
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
- **Prometheus/Alertmanager:** Export analysis results as metrics and send alerts
- **Slack/Discord/Teams:** Send analysis results to chat channels
- **Helm:** Deploy via standard Helm charts
- **ArgoCD/FluxCD:** Compatible with GitOps workflows
- **MCP:** Server integration for AI assistants
- **Custom analyzers:** HTTP-based external analyzers for extensibility

### Design

**Explain the design principles and best practices the project is following.**

- **Vendor neutrality:** Support for 13+ AI backends (OpenAI, Azure, Cohere, Ollama, Amazon Bedrock, Google Gemini, etc.)
- **Read-only operation:** Never modifies cluster state, reducing risk
- **Anonymization:** Sensitive data masked before sending to AI backends
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

k8sgpt has no in-cluster service dependencies. It reads directly from the Kubernetes API server. Optional integrations (Prometheus, Slack, custom analyzers) are external services that the user configures.

**Describe how the project implements Identity and Access Management.**

- Uses the user's existing kubeconfig for cluster authentication (in-cluster or local)
- No additional IAM roles or service accounts required
- Operator runs as a standard Kubernetes deployment with RBAC permissions for read-only access to cluster resources
- AI backend API keys are managed by the user and stored locally

**Describe how the project has addressed sovereignty.**

- Local AI options (Ollama, LocalAI) keep data within the user's network
- Users can configure any AI backend including self-hosted models
- Anonymization feature masks sensitive data before sending to external AI providers
- Remote caching (S3, Azure Blob, GCS) can be configured with user-controlled storage

**Describe any compliance requirements addressed by the project.**

- Apache 2.0 license compliance
- CNCF Code of Conduct
- OpenSSF Best Practices badge
- Developer Certificate of Origin (DCO) enforced on all commits
- FOSSA license scanning in CI

**Describe the project's High Availability requirements.**

k8sgpt is a stateless tool:
- **CLI:** Single-process, no HA requirements
- **Operator:** Standard Kubernetes deployment with configurable replicas. No leader election needed for basic operation.
- **MCP server:** Stateless HTTP server, can be horizontally scaled

**Describe the project's resource requirements, including CPU, Network and Memory.**

- **CLI:** ~50MB memory, minimal CPU (analysis time depends on cluster size and AI backend response time)
- **Operator:** ~100MB memory, minimal CPU (periodic scanning)
- **Network:** Outbound HTTPS to AI backend (when using --explain flag), outbound to Kubernetes API
- **Storage:** Config file (~1KB), optional remote caching (user-configured)

**Describe the project's storage requirements, including its use of ephemeral and/or persistent storage.**

- **Ephemeral:** No persistent storage required. Config stored locally.
- **Optional:** Remote caching (S3, Azure Blob, GCS) for analysis results
- **Operator:** No persistent volumes required

**Please outline the project's API Design:**

- **Kubernetes API:** Read-only access to standard Kubernetes resources (Pods, Deployments, Services, etc.)
- **AI Backend API:** Standard REST/HTTP API following OpenAI chat completion format
- **Custom Analyzer API:** gRPC-based protocol defined in [schema](https://github.com/k8sgpt-ai/schemas/blob/main/protobuf/schema/v1/custom_analyzer.proto)
- **MCP Server:** Model Context Protocol for AI assistant integration
- **Release API:** Automated via release-please and GoReleaser

**Describe the project's release processes, including major, minor and patch releases.**

Documented in [RELEASE.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/RELEASE.md):
- Automated via release-please (tracks conventional commits) and GoReleaser
- Monthly release cadence
- Semantic versioning (MAJOR.MINOR.PATCH)
- Binaries, container images, Helm charts, Homebrew packages, RPM/DEB/APK packages

### Installation

**Describe how the project is installed and initialized, e.g. a minimal install with a few lines of code or does it require more complex integration and configuration?**

- **CLI:** `brew install k8sgpt` or download binary — requires only an AI API key (`k8sgpt auth add`)
- **Operator:** `helm install k8sgpt-operator k8sgpt-ai/k8sgpt` — requires AI backend configuration
- **MCP:** `k8sgpt serve --mcp` — zero configuration needed

**How does an adopter test and validate the installation?**

Run `k8sgpt analyze --explain` to verify the tool connects to the cluster and produces analysis results. The operator can be validated by checking Pod status and Prometheus metrics.

### Security

**Please provide a link to the project's cloud native security self assessment.**

[SECURITY_SELF_ASSESSMENT.md](https://github.com/k8sgpt-ai/k8sgpt/blob/main/SECURITY_SELF_ASSESSMENT.md)

**Please review the Cloud Native Security Tenets from TAG Security.**

k8sgpt satisfies these cloud native security tenets:
- **Secure by default:** Anonymization is enabled by default, read-only access is the default mode
- **Least privilege:** Only reads cluster state, no write operations
- **Defense in depth:** Multiple layers of security (anonymization, TLS, RBAC, DCO, OpenSSF badge)
- **Transparency:** All code is open source, security policy is documented

**Describe how each of the cloud native principles apply to your project.**

- **Secure by default:** Sensitive data is masked before AI queries
- **Least privilege:** Read-only Kubernetes API access
- **Zero trust:** Each AI backend call is authenticated independently
- **Defense in depth:** Multiple security layers (anonymization, TLS, RBAC, DCO, OpenSSF)

**How do you recommend users alter security defaults in order to "loosen" the security of the project?**

Users who need to send unmasked data to AI backends can disable anonymization. Users who need write access should use k8sgpt as a read-only tool and implement remediation through separate Kubernetes admission controllers or operators.

**Security Hygiene**

- **Frameworks/practices:** Go modules with strict versioning, Renovate for automated dependency updates, golangci-lint, DCO enforcement, CODEOWNERS for PR review
- **Security risk evaluation:** All analyzers are reviewed for data exposure. Known risks (event message masking) are tracked in GitHub issues.

**Cloud Native Threat Modeling**

- **Least privilege:** k8sgpt only reads from the Kubernetes API. No write operations, no cluster modifications.
- **Certificate rotation:** Not applicable — k8sgpt does not manage certificates. Uses standard Kubernetes API TLS.
- **Secure software supply chain:** SBOM generated via Syft in release workflow. Renovate for automated dependency updates. OpenSSF Best Practices badge.

## Day 1 - Installation and Deployment Phase

### Project Installation and Configuration

**Describe what project installation and configuration look like.**

- **CLI:** `brew install k8sgpt` → `k8sgpt auth add --backend openai` → `k8sgpt analyze --explain`
- **Operator:** Helm chart with configurable AI backends, analyzers, and integrations
- **Custom analyzers:** gRPC-based HTTP service + k8sgpt config

### Project Enablement and Rollback

**How can this project be enabled or disabled in a live cluster?**

- **CLI:** No enable/disable needed — run when needed
- **Operator:** `helm upgrade --set enabled=false` or `helm uninstall`

**Describe how enabling the project changes any default behavior of the cluster or running workloads.**

k8sgpt is purely read-only. It has no impact on cluster behavior or running workloads.

**Describe how the project tests enablement and disablement.**

CI tests cover both CLI and operator modes. The operator is tested with Minikube and Kind clusters.

**How does the project clean up any resources created, including CRDs?**

The operator creates no persistent resources. Helm uninstall removes all resources. CLI creates no cluster resources.

### Rollout, Upgrade and Rollback Planning

**How does the project intend to provide and maintain compatibility with infrastructure and orchestration management tools like Kubernetes and with what frequency?**

- Uses client-go for Kubernetes API interaction
- Tests against current and previous 2 Kubernetes versions
- Monthly release cadence ensures timely updates

**Describe how the project handles rollback procedures.**

- **CLI:** Rollback via `brew upgrade k8sgpt@previous` or downloading previous binary
- **Operator:** `helm rollback` to previous release
- **Container images:** Rollback via image tag

**How can a rollout or rollback fail? Describe any impact to already running workloads.**

No impact on running workloads — k8sgpt is read-only. Rollback failures would only affect k8sgpt's ability to function, not cluster operations.

**Describe any specific metrics that should inform a rollback.**

- Operator crash loops
- Analysis result quality degradation
- AI backend API errors

**Explain how upgrades and rollbacks were tested.**

Each release includes CI tests against multiple Kubernetes versions. Operator tests run in Kind and Minikube environments.

**Explain how the project informs users of deprecations and removals of features and APIs.**

- Changelog in GitHub Releases
- Deprecation warnings in CLI output
- Documentation updates on docs.k8sgpt.ai

**Explain how the project permits utilization of alpha and beta capabilities as part of a rollout.**

k8sgpt uses only stable Kubernetes API versions. Custom analyzers can leverage alpha APIs if users configure them to do so.

## Day 2 - Day-to-Day Operations Phase

### Scalability/Reliability

**Describe how the project increases the size or count of existing API objects.**

Not applicable — k8sgpt is a read-only analysis tool, not a controller that manages API objects.

**Describe how the project defines Service Level Objectives (SLOs) and Service Level Indicators (SLIs).**

- **SLO:** Analysis completion within expected timeframes (varies by cluster size and AI backend)
- **SLI:** Percentage of successful analysis runs, error rates, AI backend response times

**Describe any operations that will increase in time covered by existing SLIs/SLOs.**

Larger clusters and more analyzers increase analysis time proportionally. The operator can be configured to scan namespaces independently.

**Describe the increase in resource usage in any components as a result of enabling this project.**

- **CLI:** ~50MB memory, minimal CPU
- **Operator:** ~100MB memory, minimal CPU
- **Network:** Outbound HTTPS to AI backend per analysis

**Describe which conditions enabling / using this project would result in resource exhaustion.**

Not applicable — k8sgpt is read-only and does not create persistent resources.

**Describe the load testing that has been performed on the project and the results.**

Load testing is performed by the community during development. The project scales to clusters with 1000+ nodes. Operator scanning is configurable for different intervals.

**Describe the recommended limits of users, requests, system resources, etc.**

No hard limits. The CLI works with any cluster size. The operator is recommended for clusters with 100+ nodes where continuous monitoring is valuable.

**Describe which resilience pattern the project uses.**

- **CLI:** Stateless, no circuit breaker needed
- **Operator:** Standard Kubernetes deployment with health checks
- **AI backends:** Multiple backend support provides natural fallback

### Observability Requirements

**Describe the signals the project is using or producing, including logs, metrics, profiles and traces.**

- **Logs:** Standard Kubernetes operator logging (info, warn, error levels)
- **Metrics:** Prometheus-compatible metrics for operator mode (analysis duration, error rates)
- **Traces:** Not currently implemented, but the architecture supports it
- **Formats:** JSON logs, Prometheus exposition format

**Describe how the project captures audit logging.**

k8sgpt does not modify cluster state, so audit logging is not applicable. The Kubernetes API server audit logs capture all k8sgpt read operations.

**Describe any dashboards the project uses or implements.**

k8sgpt does not include built-in dashboards, but integrates with:
- **Prometheus:** Metrics can be visualized in Grafana
- **Custom integrations:** Slack, Teams for alerting

**Describe how the project surfaces project resource requirements for adopters to monitor cloud and infrastructure costs.**

k8sgpt has minimal resource requirements (~100MB memory for operator). No persistent storage needed. Cost tracking is via standard Kubernetes resource monitoring.

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
- Slack/Teams alert channels

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

- Automated via Renovate with auto-merge for non-major updates
- Go modules with strict version pinning
- Monthly release cadence ensures timely dependency updates
- Security vulnerabilities tracked via OpenSSF badge and FOSSA

**How does the project incorporate and consider source composition analysis as part of its development and security hygiene?**

- FOSSA license scanning in CI
- Renovate for automated dependency tracking
- SBOM generated via Syft in release workflow
- OpenSSF Best Practices badge

**Describe how the project implements changes based on source composition analysis (SCA) and the timescale.**

- FOSSA findings are reviewed and addressed within the next release cycle
- Critical security vulnerabilities in dependencies are patched immediately
- Non-critical license issues are addressed within 30 days

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

- Apache 2.0 license file in repository root
- LICENSE in all subdirectories
- FOSSA license scanning in CI
- Go modules with license metadata
- Copyright headers in source files

**Describe how the project ensures alignment with CNCF recommendations for attribution notices.**

- Standard Go module attribution
- LICENSE files in all packages
- FOSSA license compliance scanning
- Copyright headers in source files following Apache 2.0 conventions

## Day 2 - Security

### Security Hygiene

**How is the project executing access control?**

- GitHub CODEOWNERS enforces PR review requirements
- Maintainer team (7 members) with clear roles and responsibilities
- DCO enforcement on all commits
- Branch protection requires PR review and passing CI
- GitHub organization membership controls write access

### Cloud Native Threat Modeling

**How does the project ensure its security reporting and response team is representative of its community diversity (organizational and individual)?**

The maintainer team includes contributors from:
- AWS (AlexsJones)
- Agicap (matthisholleville)
- DaoCloud (yankay)
- @basiqio (bradmccoydev)
- Independent contributors (thschue, AnaisUrlichs, roberthstrand, rakshitgondwal)

**How does the project invite and rotate security reporting team members?**

Security reports are handled via:
- Email: contact@k8sgpt.ai
- Slack: Any maintainer in #k8sgpt
- GitHub Security Advisories

Any maintainer can respond to security reports. The process is documented in SECURITY.md and the security self-assessment.
