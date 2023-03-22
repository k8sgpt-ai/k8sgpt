<img src="images/logo.png" width="100px;" />

# k8sgpt

<br />
AI Powered Kubernetes debugging for SRE, Platform and DevOps teams.
<br />

<img src="images/demo.gif" width=650px; />

## What is k8sgpt?

`k8sgpt` is a tool for scanning your kubernetes clusters, diagnosing and triaging issues in simple english.

It reduces the mystery of kubernetes and makes it easy to understand what is going on in your cluster.

## Usage

```
# Ensure KUBECONFIG env is set to an active Kubernetes cluster
k8sgpt auth key <Your OpenAI key>

k8sgpt find problems --explain

```

### What about kubectl-ai?

The the kubectl-ai [project](https://github.com/sozercan/kubectl-ai) uses AI to create manifests and apply them to the cluster. It is not what we are trying to do here, it is focusing on writing YAML manifests.

K8sgpt is focused on triaging and diagnosing issues in your cluster. It is a tool for SRE, Platform & DevOps engineers to help them understand what is going on in their cluster. Cutting through the noise of logs and multiple tools to find the root cause of an issue.


### Configuration 

`k8sgpt` stores config data in `~/.k8sgpt` the data is stored in plain text, including your OpenAI key.

### Community
* Find us on [Slack](https://cloud-native.slack.com/channels/k8sgpt-ai)
