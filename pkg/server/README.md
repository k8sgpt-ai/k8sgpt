# serve

The serve commands allow you to run k8sgpt in a grpc server mode.
This would be enabled typically through `k8sgpt serve` and is how the in-cluster k8sgpt deployment functions when managed by the [k8sgpt-operator](https://github.com/k8sgpt-ai/k8sgpt-operator)

The grpc interface that is served is hosted on [buf](https://buf.build/k8sgpt-ai/schemas) and the repository for this is [here](https://github.com/k8sgpt-ai/schemas)

## grpcurl

A fantastic tool for local debugging and development is `grpcurl` 
It allows you to form curl like requests that are http2
e.g. 

```
grpcurl -plaintext -d '{"namespace": "k8sgpt", "explain" : "true"}' localhost:8080 schema.v1.ServerService/Analyze
```

```
grpcurl -plaintext  localhost:8080 schema.v1.ServerService/ListIntegrations 
{
  "integrations": [
    "trivy"
  ]
}

```

```
grpcurl -plaintext -d '{"integrations":{"trivy":{"enabled":"true","namespace":"default","skipInstall":"false"}}}' localhost:8080 schema.v1.ServerService/AddConfig
```
