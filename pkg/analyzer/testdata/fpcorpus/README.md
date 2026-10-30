# False-positive regression corpus

Each file in this directory is a Kubernetes scenario that an analyzer once
reported incorrectly (or, for `*-positive.yaml` controls, one it must keep
reporting). `TestFalsePositiveCorpus` in `pkg/analyzer/fp_corpus_test.go`
loads every case, runs the named analyzer against a fake clientset seeded with
the objects, and compares the reported result names with `expectResults`.

## Layout

```
testdata/fpcorpus/<Analyzer>/<case>.yaml
```

`<Analyzer>` is the analyzer name as registered in `analyzer.go` (for example
`Node`, `Job`, `HorizontalPodAutoscaler`).

## Case format

The first YAML document describes the case. The documents after it are the
Kubernetes objects to seed the fake clientset with.

```yaml
issues: ["#1461"]                 # issues/PRs this case guards
description: EKS node monitoring agent conditions are healthy when True
namespace: ""                     # namespace passed to the analyzer ("" = all)
expectResults: []                 # result names the analyzer must report
---
apiVersion: v1
kind: Node
...
```

An empty `expectResults` means the analyzer must report nothing for the
scenario.

## Adding a case

When you fix a false positive, add a case here that reproduces the original
report. The test must fail on the code before your fix and pass after it.
Prefer the smallest set of objects that reproduces the problem.

Only objects the client-go scheme knows (core and built-in API groups) can be
used here today. Analyzers that need CRDs or a dynamic client still use their
own unit tests.
