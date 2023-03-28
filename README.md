<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./images/banner-white.png" width="600px;">
  <img alt="Text changing depending on mode. Light: 'So light!' Dark: 'So dark!'" src="./images/banner-black.png" width="600px;">
</picture>

`k8sgpt` is a tool for scanning your kubernetes clusters, diagnosing and triaging issues in simple english.

It has SRE experience codified into it's analyzers and helps to pull out the most relevent information to enrich it with AI.

## Quick Start

```
brew tap k8sgpt-ai/k8sgpt
brew install k8sgpt
```

* Currently the default AI provider is OpenAI, you will need to generate an API key from [OpenAI](https://openai.com)
  * You can do this by running `k8sgpt generate` to open a browser link to generate it 
* Run `k8sgpt auth` to set it in k8sgpt.
* Run `k8sgpt analyze` to run a scan.
* And use `k8sgpt analyze --explain` to get a more detailed explanation of the issues.

<img src="images/demo4.gif" width=650px; />

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
Usage:
  k8sgpt [command]

Available Commands:
  analyze     This command will find problems within your Kubernetes cluster
  auth        Authenticate with your chosen backend
  completion  Generate the autocompletion script for the specified shell
  generate    Generate Key for your chosen backend (opens browser)
  help        Help about any command
  version     Print the version number of k8sgpt

Flags:
      --config string       config file (default is $HOME/.k8sgpt.git.yaml)
  -h, --help                help for k8sgpt
      --kubeconfig string   Path to a kubeconfig. Only required if out-of-cluster.
      --master string       The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.
  -t, --toggle              Help message for toggle

Use "k8sgpt [command] --help" for more information about a command.
```

_Run a scan with the default analyzers_

```
k8sgpt generate
k8sgpt auth
k8sgpt analyze --explain
```

_Filter on resource_

```
k8sgpt analyze --explain --filter=Service
```

_Output to JSON_

```
k8sgpt analyze --explain --filter=Service --output=json
```

## Upcoming major milestones

- [ ] Multiple AI backend support
- [ ] Custom AI/ML model backend support
- [ ] Custom analyzers

### What about kubectl-ai?

The the kubectl-ai [project](https://github.com/sozercan/kubectl-ai) uses AI to create manifests and apply them to the cluster. It is not what we are trying to do here, it is focusing on writing YAML manifests.

K8sgpt is focused on triaging and diagnosing issues in your cluster. It is a tool for SRE, Platform & DevOps engineers to help them understand what is going on in their cluster. Cutting through the noise of logs and multiple tools to find the root cause of an issue.


### Configuration 

`k8sgpt` stores config data in `~/.k8sgpt.yaml` the data is stored in plain text, including your OpenAI key.

### Contributing

Please read our [contributing guide](./CONTRIBUTING.md).
### Community
* Find us on [Slack](https://cloud-native.slack.com/channels/k8sgpt-ai)
