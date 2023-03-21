<img src="images/logo.png" width="100px;" />

# k8sgpt

<br />
AI Powered Kubernetes debugging for SRE, Platform and DevOps teams.
<br />

<img src="images/demo.gif" width=800px; />

## What is k8sgpt?

`k8sgpt` is a tool for scanning your kubernetes clusters, diagnosing and triaging issues in simple english.
It reduces the mystery of kubernetes and makes it easy to understand what is going on in your cluster.


## Usage

```
k8sgpt auth key <Your OpenAI key>

k8sgpt find problems --explain

```


### Configuration 

`k8sgpt` stores config data in `~/.k8sgpt` the data is stored in plain text, including your OpenAI key.