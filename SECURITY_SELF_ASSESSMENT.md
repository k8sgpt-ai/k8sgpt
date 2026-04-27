# k8sgpt Self-assessment

## Metadata

| Field | Value |
|-------|-------|
| Assessment Stage | Incomplete |
| Software | https://github.com/k8sgpt-ai/k8sgpt |
| Security Provider | No |
| Languages | Go |
| SBOM | Generated via Syft in release workflow (https://github.com/k8sgpt-ai/k8sgpt/blob/main/.github/workflows/release.yaml) |

### Security Links

| Doc | URL |
|-----|-----|
| Security Policy | https://github.com/k8sgpt-ai/k8sgpt/blob/main/SECURITY.md |
| Code of Conduct | https://github.com/k8sgpt-ai/k8sgpt/blob/main/CODE_OF_CONDUCT.md |
| OpenSSF Best Practices | https://bestpractices.coreinfrastructure.org/projects/7272 |
| Contributing Guide | https://github.com/k8sgpt-ai/k8sgpt/blob/main/CONTRIBUTING.md |
| Documentation | https://docs.k8sgpt.ai |

## Overview

k8sgpt is a tool for scanning Kubernetes clusters, diagnosing, and triaging issues in simple English. It has SRE experience codified into analyzers and helps pull out the most relevant information, enriching it with AI from various LLM providers.

### Background

k8sgpt operates by connecting to a Kubernetes cluster, running a suite of analyzers that check for common issues (pod failures, misconfigurations, resource limits, etc.), and using AI to explain findings in plain language. It is designed as a CLI tool and as a Kubernetes operator for continuous monitoring.

### Actors

1. **k8sgpt CLI** - The command-line tool that connects to a Kubernetes cluster via kubeconfig, runs analyzers, and displays results.
2. **k8sgpt Operator** - A Kubernetes controller that continuously monitors the cluster and reports issues via alerts.
3. **AI Backend** - External LLM provider (OpenAI, Azure, Cohere, Ollama, Amazon Bedrock, etc.) that provides analysis explanations. The operator does not communicate with external APIs.
4. **Kubernetes Cluster** - The target cluster being analyzed. k8sgpt reads cluster state via the Kubernetes API.
5. **Custom Analyzer** - Optional external HTTP service that can be called for custom analysis (via the custom analyzer framework).

### Actions

1. **Connection**: k8sgpt authenticates to the Kubernetes cluster using the user's kubeconfig (in-cluster or local).
2. **Analysis**: Analyzers query the Kubernetes API for specific resource states and check for known issue patterns.
3. **Filtering**: Results are filtered based on user-configured filters to reduce noise.
4. **AI Explanation** (optional): When `--explain` is used, analysis results are sent to an AI backend for natural language explanation. Sensitive data is masked before being sent.
5. **Output**: Results are displayed in the CLI, exported to JSON, or sent via integrations (Slack, Prometheus, etc.).
6. **Operator Mode**: The k8sgpt operator runs continuously, performing periodic scans and alerting on issues.

### Goals

- Provide accessible Kubernetes diagnostics for users of all skill levels
- Codify SRE expertise into automated analyzers
- Support multiple AI backends for flexibility and vendor neutrality
- Enable continuous cluster monitoring via the operator
- Maintain data privacy through anonymization of sensitive cluster data before sending to AI backends

### Non-goals

- k8sgpt does not modify or remediate cluster state (it is read-only)
- k8sgpt does not replace comprehensive security scanning tools (it focuses on Kubernetes operational issues)
- k8sgpt does not store cluster data persistently (except optional remote caching)
- k8sgpt does not provide its own AI model (it consumes external AI APIs)

## Self-assessment use

This self-assessment is created by the k8sgpt team to perform an internal analysis of the project's security. It is not intended to provide a security audit of k8sgpt, or function as an independent assessment or attestation of k8sgpt's security health.

This document serves to provide k8sgpt users with an initial understanding of k8sgpt's security, where to find existing security documentation, k8sgpt plans for security, and general overview of k8sgpt security practices, both for development of k8sgpt as well as security of k8sgpt.

This document provides the CNCF TAG-Security with an initial understanding of k8sgpt to assist in a joint-assessment, necessary for projects under incubation. Taken together, this document and the joint-assessment serve as a cornerstone for if and when k8sgpt seeks graduation and is preparing for a security audit.

## Security Functions and Features

### Critical

| Component | Description |
|-----------|-------------|
| **Data Anonymization** | k8sgpt masks sensitive Kubernetes resource data (names, labels, etc.) before sending to AI backends. This prevents accidental exposure of sensitive cluster state. |
| **Read-only API Access** | k8sgpt only reads from the Kubernetes API. It never modifies cluster state, reducing the attack surface. |
| **Secure Config Storage** | AI API keys are stored locally in `$XDG_CONFIG_HOME/k8sgpt/k8sgpt.yaml`. Users are responsible for securing this file. |
| **TLS for API Communication** | All communication with AI backends uses HTTPS/TLS. |

### Security Relevant

| Component | Description |
|-----------|-------------|
| **AI Backend Configuration** | Users can configure which AI backend to use. Local options (Ollama, LocalAI) keep data within the user's network. |
| **Anonymization Settings** | Users can enable/disable anonymization. When enabled, sensitive data is masked before AI queries. |
| **Filter Configuration** | Users can configure which analyzers and filters are active, controlling what data is processed. |
| **Remote Caching** | Optional S3/Azure/GCS caching. Users must secure their storage credentials. |
| **Custom Headers** | Users can add custom headers to AI requests for additional authentication (e.g., API keys). |
| **Container Image Signing** | Release workflow generates SBOM via Syft for supply chain transparency. |

## Project Compliance

k8sgpt currently complies with:
- **CNCF Code of Conduct** - All contributors must follow the CNCF CoC
- **OpenSSF Best Practices** - Project maintains an OpenSSF Best Practices badge
- **Apache 2.0 License** - Clear, permissive open source license
- **DCO (Developer Certificate of Origin)** - All commits must be signed off

## Secure Development Practices

### Development Pipeline

- **Go Modules** - All dependencies are managed via Go modules with strict version pinning
- **Renovate** - Automated dependency updates with auto-merge for non-major versions
- **CI Pipeline** - Every PR triggers:
  - Go compilation and build
  - Unit and integration tests
  - Go linter (golangci-lint)
  - Semantic PR validation
  - Container image build
- **Code Review** - All changes require at least one maintainer review (enforced via CODEOWNERS)
- **DCO Enforcement** - All commits must be signed off via DCO check
- **Branch Protection** - Main branch requires PR review and passing CI
- **Conventional Commits** - All commits follow conventional commit format for release automation
- **Release Process** - Automated releases via release-please and GoReleaser

### Communication Channels

- **Internal**: GitHub Issues, GitHub Discussions, Slack (#k8sgpt)
- **Inbound**: GitHub Issues (bug reports, feature requests), Slack (#k8sgpt), Email (contact@k8sgpt.ai for security)
- **Outbound**: GitHub Releases, Slack announcements, CNCF mailing list, blog posts

### Ecosystem

k8sgpt is deeply integrated into the Kubernetes ecosystem:
- Native Kubernetes API interaction via client-go
- Helm chart distribution via k8sgpt-ai/charts
- Krew plugin distribution
- Homebrew tap for easy installation
- Integration with Prometheus and Alertmanager for monitoring
- Compatible with GitOps tools (ArgoCD, FluxCD)
- MCP server integration for AI assistant workflows

## Security Issue Resolution

### Responsible Disclosure Process

Users who discover security vulnerabilities in k8sgpt should:

1. **Report via Email**: Send details to contact@k8sgpt.ai
2. **Report via Slack**: Contact a maintainer in the #k8sgpt Slack channel
3. **GitHub Security Advisories**: Use the "Report a vulnerability" button on the GitHub repository

### Vulnerability Response Process

1. **Acknowledgment**: Maintainers acknowledge receipt within 48 hours
2. **Assessment**: The vulnerability is assessed for severity and impact
3. **Fix**: A fix is developed and tested
4. **Disclosure**: The vulnerability is disclosed via GitHub Security Advisories and a new release is published
5. **Communication**: Users are notified via GitHub Releases and Slack

### Incident Response

1. **Triage**: Security reports are triaged by the maintainer team
2. **Confirmation**: The reported issue is confirmed and severity is assessed
3. **Notification**: Affected users are notified via GitHub and Slack
4. **Patching**: A fix is developed and published in a patch release
5. **Post-mortem**: If applicable, a post-mortem is shared with the community

### Known Security Considerations

- **AI Backend Data Exposure**: While anonymization is the default, users should be aware that analysis results (including Kubernetes resource names) may be sent to external AI providers when using `--explain`. Users concerned about data privacy should use local models (Ollama, LocalAI) or review the anonymization settings.
- **Config File Security**: AI API keys are stored in plaintext in the local config file. Users should secure this file and restrict access.
- **Custom Analyzer Security**: Custom analyzers run as separate HTTP services. Users should ensure these services are properly secured and only accessible from trusted networks.

## Appendix

### OpenSSF Best Practices

k8sgpt maintains an [OpenSSF Best Practices badge](https://bestpractices.coreinfrastructure.org/projects/7272). The project actively addresses non-passing criteria and works toward a 100% score.

### Related Projects

- **Prometheus** - Prometheus provides metrics collection; k8sgpt provides diagnostic analysis
- **kube-linter** - kube-linter focuses on static analysis of Kubernetes manifests; k8sgpt provides runtime analysis
- **Sonobuoy** - Sonobuoy provides Kubernetes cluster conformance testing; k8sgpt provides operational issue detection
- **Kubescape** - Kubescape focuses on security posture assessment; k8sgpt focuses on operational diagnostics with AI-powered explanations

### Case Studies

1. **Day-to-day cluster troubleshooting**: SREs use k8sgpt to quickly identify and explain common Kubernetes issues (crashlooping pods, failed deployments, resource limits) without deep domain knowledge.
2. **On-call support**: Junior team members use k8sgpt's AI-powered explanations to understand and resolve cluster issues during on-call rotations.
3. **Continuous monitoring**: Teams deploy the k8sgpt operator to continuously monitor cluster health and alert on issues via Prometheus/Alertmanager integration.
