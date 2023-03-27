<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./images/banner-white.png" width="600px;">
  <img alt="Text changing depending on mode. Light: 'So light!' Dark: 'So dark!'" src="./images/banner-black.png" width="600px;">
</picture>

_Install it now_

```
brew tap k8sgpt-ai/k8sgpt
brew install k8sgpt
```

`k8sgpt` is a tool for scanning your kubernetes clusters, diagnosing and triaging issues in simple english.

It has SRE experience codified into it's analyzers and helps to pull out the most relevent information to enrich it with AI.

<img src="images/landing.png" width=650px; />

## Analyzers

K8sGPT uses analyzers to triage and diagnose issues in your cluster. It has a set of analyzers that are built in, but you will be able to write your own analyzers.

### Built in analyzers

- [x] podAnalyzer
- [x] pvcAnalyzer
- [x] rsAnalyzer
- [x] serviceAnalyzer
- [x] eventAnalyzer

## Usage

```
Kubernetes debugging powered by AI

Usage:
  k8sgpt [command]

Available Commands:
  analyze     This command will find problems within your Kubernetes cluster
  auth        Authenticate with your chosen backend
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command

```

_Run a scan with the default analyzers_

```
k8sgpt auth
k8sgpt analyze --explain
```

_Filter on resource_

```
k8sgpt analyze --explain --resource=Service
```

_Output to JSON_

```
k8sgpt analyze --explain --resource=Service --output=json
```

## Upcoming major milestones

- [ ] Multiple AI backend support
- [ ] Custom AI/ML model backend support
- [ ] Custom analyzers

### What about kubectl-ai?

The the kubectl-ai [project](https://github.com/sozercan/kubectl-ai) uses AI to create manifests and apply them to the cluster. It is not what we are trying to do here, it is focusing on writing YAML manifests.

K8sgpt is focused on triaging and diagnosing issues in your cluster. It is a tool for SRE, Platform & DevOps engineers to help them understand what is going on in their cluster. Cutting through the noise of logs and multiple tools to find the root cause of an issue.


### Configuration 

`k8sgpt` stores config data in `~/.k8sgpt` the data is stored in plain text, including your OpenAI key.

### Contributing

Please read our [contributing guide](./CONTRIBUTING.md).
### Community
* Find us on [Slack](https://cloud-native.slack.com/channels/k8sgpt-ai)
