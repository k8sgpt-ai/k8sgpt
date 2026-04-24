# k8sgpt Roadmap

This document outlines the strategic direction and roadmap for k8sgpt. The roadmap is updated regularly based on community feedback and project needs.

## Project Vision

k8sgpt aims to give Kubernetes superpowers to everyone by providing AI-powered cluster analysis that is accessible to users of all skill levels. We believe every Kubernetes practitioner should have access to SRE-grade diagnostics, regardless of their experience.

## Current Focus Areas

### 1. Multi-Provider AI Backend Support
- Expand support for diverse LLM providers beyond OpenAI
- Support for local/offline models (Ollama, LocalAI) as first-class citizens
- Improved model selection and fallback mechanisms
- Support for custom REST backends

### 2. Analyzer Expansion
- Add analyzers for emerging Kubernetes features
- Improve existing analyzer accuracy
- Custom analyzer framework improvements
- Integration with OPA/Gatekeeper for policy analysis

### 3. Observability & Integration
- Deeper Prometheus and OpenTelemetry integration
- Grafana dashboard support
- Slack, Discord, and Teams integrations
- Webhook-based alerting

### 4. Performance & Scalability
- Optimize analysis speed for large clusters
- Improve caching mechanisms
- Reduce memory footprint
- Support for multi-cluster analysis

### 5. Security
- Enhanced data anonymization during AI queries
- Improved security analyzers
- Compliance reporting features
- Secret detection and redaction

## Planned Initiatives

### Near Term (Next 3-6 months)
- [ ] Kubernetes 1.30+ feature support
- [ ] Improved MCP server capabilities
- [ ] Enhanced serve mode with metrics
- [ ] Better error handling and diagnostics
- [ ] Community analyzer marketplace

### Medium Term (6-12 months)
- [ ] Multi-cluster analysis
- [ ] Advanced anomaly detection
- [ ] Historical trend analysis
- [ ] Custom dashboard generation
- [ ] Plugin ecosystem for third-party analyzers

### Long Term (12+ months)
- [ ] Self-healing recommendations with automated remediation
- [ ] Cross-cluster best practice benchmarking
- [ ] AI model fine-tuning for Kubernetes diagnostics
- [ ] Enterprise-grade access control and auditing

## How to Contribute

The roadmap is community-driven. To suggest new items:
1. Open a GitHub issue with the `enhancement` label
2. Discuss your idea in the [#k8sgpt Slack channel](https://join.slack.com/t/k8sgpt/shared_invite/zt-332vhyaxv-bfjJwHZLXWVCB3QaXafEYQ)
3. Submit a PR implementing the feature

## Previous Roadmaps

- _(Initial roadmap created 2026-04-24)_

## Feedback

We welcome feedback on this roadmap. Please share your thoughts via:
- GitHub Issues
- Slack: [#k8sgpt](https://join.slack.com/t/k8sgpt/shared_invite/zt-332vhyaxv-bfjJwHZLXWVCB3QaXafEYQ)
- GitHub Discussions
