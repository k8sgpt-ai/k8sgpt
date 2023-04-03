<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./images/banner-white.png" width="600px;">
  <img alt="Text changing depending on mode. Light: 'So light!' Dark: 'So dark!'" src="./images/banner-black.png" width="600px;">
</picture>
<br/>

![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/k8sgpt-ai/k8sgpt)
![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/k8sgpt-ai/k8sgpt/release.yaml)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/k8sgpt-ai/k8sgpt)

`k8sgpt` is a tool for scanning your Kubernetes clusters, diagnosing, and triaging issues in simple English.

It has SRE experience codified into its analyzers and helps to pull out the most relevant information to enrich it with AI.

## Installation

### Linux/Mac via brew

```
brew tap k8sgpt-ai/k8sgpt
brew install k8sgpt
```

<details>
  <summary>Failing Installation on WSL or Linux (missing gcc)</summary>
  When installing Homebrew on WSL or Linux, you may encounter the following error:

  ```
  ==> Installing k8sgpt from k8sgpt-ai/k8sgpt Error: The following formula cannot be installed from a bottle and must be 
  built from the source. k8sgpt Install Clang or run brew install gcc.
  ```

If you install gcc as suggested, the problem will persist. Therefore, you need to install the build-essential package.
  ```
     sudo apt-get update
     sudo apt-get install build-essential
  ```
</details>

### Windows

* Download the latest Windows binaries of **k8sgpt** from the [Release](https://github.com/k8sgpt-ai/k8sgpt/releases) 
  tab based on your system architecture.
* Extract the downloaded package to your desired location. Configure the system *path* variable with the binary location

### Verify installation

* Run `k8sgpt version`

## Quick Start

* Currently the default AI provider is OpenAI, you will need to generate an API key from [OpenAI](https://openai.com)
  * You can do this by running `k8sgpt generate` to open a browser link to generate it
* Run `k8sgpt auth` to set it in k8sgpt. 
  * You can provide the password directly using the `--password` flag.
* Run `k8sgpt filters` to manage the active filters used by the analyzer. By default, all filters are executed during analysis.
* Run `k8sgpt analyze` to run a scan.
* And use `k8sgpt analyze --explain` to get a more detailed explanation of the issues.

<img src="images/demo4.gif" width=650px; />

## Analyzers

K8sGPT uses analyzers to triage and diagnose issues in your cluster. It has a set of analyzers that are built in, but 
you will be able to write your own analyzers.

### Built in analyzers

#### Enabled by default

- [x] podAnalyzer
- [x] pvcAnalyzer
- [x] rsAnalyzer
- [x] serviceAnalyzer
- [x] eventAnalyzer
- [x] ingressAnalyzer

#### Optional

- [x] hpaAnalyzer
- [x] pdbAnalyzer

## Usage

```
Usage:
  k8sgpt [command]

Available Commands:
  analyze     This command will find problems within your Kubernetes cluster
  auth        Authenticate with your chosen backend
  completion  Generate the autocompletion script for the specified shell
  filters     Manage filters for analyzing Kubernetes resources
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

_Manage filters_

_List filters_

```
k8sgpt filters list
```

_Add default filters_

```
k8sgpt filters add [filter(s)]
```

### Examples :

- Simple filter : `k8sgpt filters add Service`
- Multiple filters : `k8sgpt filters add Ingress,Pod`

_Add default filters_

```
k8sgpt filters remove [filter(s)]
```

### Examples :

- Simple filter : `k8sgpt filters remove Service`
- Multiple filters : `k8sgpt filters remove Ingress,Pod`

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

_Filter by namespace_
```
k8sgpt analyze --explain --filter=Pod --namespace=default
```

_Output to JSON_

```
k8sgpt analyze --explain --filter=Service --output=json
```

## Upcoming major milestones

- [ ] Multiple AI backend support
- [ ] Custom AI/ML model backend support
- [ ] Custom analyzers

## What about kubectl-ai?

The kubectl-ai [project](https://github.com/sozercan/kubectl-ai) uses AI to create manifests and apply them to the 
cluster. It is not what we are trying to do here, it is focusing on writing YAML manifests.

K8sgpt is focused on triaging and diagnosing issues in your cluster. It is a tool for SRE, Platform & DevOps engineers 
to help them understand what is going on in their cluster. Cutting through the noise of logs and multiple tools to find 
the root cause of an issue.


## Configuration

`k8sgpt` stores config data in `~/.k8sgpt.yaml` the data is stored in plain text, including your OpenAI key.

## Contributing

Please read our [contributing guide](./CONTRIBUTING.md).
## Community
Find us on [Slack](https://k8sgpt.slack.com/)

<a href="https://github.com/k8sgpt-ai/k8sgpt/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=k8sgpt-ai/k8sgpt" />
</a>