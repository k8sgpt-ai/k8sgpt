# Generic CRD Analyzer Configuration Examples

The Generic CRD Analyzer enables K8sGPT to automatically analyze custom resources from any installed CRD in your Kubernetes cluster. This provides observability for operator-managed resources like cert-manager, ArgoCD, Kafka, and more.

## Basic Configuration

The CRD analyzer is configured via the K8sGPT configuration file (typically `~/.config/k8sgpt/k8sgpt.yaml`). Here's a minimal example:

```yaml
crd_analyzer:
  enabled: true
```

With this basic configuration, the analyzer will:
- Discover all CRDs installed in your cluster
- Apply generic health checks based on common Kubernetes patterns
- Report issues with resources that have unhealthy status conditions

## Configuration Options

### Complete Example

```yaml
crd_analyzer:
  enabled: true
  include:
    - name: certificates.cert-manager.io
      statusPath: ".status.conditions"
      readyCondition:
        type: "Ready"
        expectedStatus: "True"
    
    - name: applications.argoproj.io
      statusPath: ".status.health.status"
      expectedValue: "Healthy"
    
    - name: kafkas.kafka.strimzi.io
      readyCondition:
        type: "Ready"
        expectedStatus: "True"
  
  exclude:
    - name: kafkatopics.kafka.strimzi.io
    - name: servicemonitors.monitoring.coreos.com
```

### Configuration Fields

#### `enabled` (boolean)
- **Default**: `false`
- **Description**: Master switch to enable/disable the CRD analyzer
- **Example**: `enabled: true`

#### `include` (array)
- **Description**: List of CRDs with custom health check configurations
- **Fields**:
  - `name` (string, required): The full CRD name (e.g., `certificates.cert-manager.io`)
  - `statusPath` (string, optional): JSONPath to the status field to check (e.g., `.status.health.status`)
  - `readyCondition` (object, optional): Configuration for checking a Ready-style condition
    - `type` (string): The condition type to check (e.g., `"Ready"`)
    - `expectedStatus` (string): Expected status value (e.g., `"True"`)
  - `expectedValue` (string, optional): Expected value at the statusPath (requires `statusPath`)

#### `exclude` (array)
- **Description**: List of CRDs to skip during analysis
- **Fields**:
  - `name` (string): The full CRD name to exclude

## Use Cases

### 1. cert-manager Certificate Analysis

Detect certificates that are not ready or have issuance failures:

```yaml
crd_analyzer:
  enabled: true
  include:
    - name: certificates.cert-manager.io
      readyCondition:
        type: "Ready"
        expectedStatus: "True"
```

**Detected Issues:**
- Certificates with `Ready=False`
- Certificate renewal failures
- Invalid certificate configurations

### 2. ArgoCD Application Health

Monitor ArgoCD application sync and health status:

```yaml
crd_analyzer:
  enabled: true
  include:
    - name: applications.argoproj.io
      statusPath: ".status.health.status"
      expectedValue: "Healthy"
```

**Detected Issues:**
- Applications in `Degraded` state
- Sync failures
- Missing resources

### 3. Kafka Operator Resources

Check Kafka cluster health with Strimzi operator:

```yaml
crd_analyzer:
  enabled: true
  include:
    - name: kafkas.kafka.strimzi.io
      readyCondition:
        type: "Ready"
        expectedStatus: "True"
  exclude:
    - name: kafkatopics.kafka.strimzi.io  # Exclude topics to reduce noise
```

**Detected Issues:**
- Kafka clusters not ready
- Broker failures
- Configuration issues

### 4. Prometheus Operator

Monitor Prometheus instances:

```yaml
crd_analyzer:
  enabled: true
  include:
    - name: prometheuses.monitoring.coreos.com
      readyCondition:
        type: "Available"
        expectedStatus: "True"
```

**Detected Issues:**
- Prometheus instances not available
- Configuration reload failures
- Storage issues

## Generic Health Checks

When a CRD is not explicitly configured in the `include` list, the analyzer applies generic health checks:

### Supported Patterns

1. **status.conditions** - Standard Kubernetes conditions
   - Flags `Ready` conditions with status != `"True"`
   - Flags any condition type containing "failed" with status = `"True"`

2. **status.phase** - Phase-based resources
   - Flags resources with phase = `"Failed"` or `"Error"`

3. **status.health.status** - ArgoCD-style health
   - Flags resources with health status != `"Healthy"` (except `"Unknown"`)

4. **status.state** - State-based resources
   - Flags resources with state = `"Failed"` or `"Error"`

5. **Deletion with Finalizers** - Stuck resources
   - Flags resources with `deletionTimestamp` set but still having finalizers

## Running the Analyzer

### Enable in Configuration

Add the CRD analyzer to your active filters:

```bash
# Add CustomResource filter
k8sgpt filters add CustomResource

# List active filters to verify
k8sgpt filters list
```

### Run Analysis

```bash
# Basic analysis
k8sgpt analyze --explain

# With specific filter
k8sgpt analyze --explain --filter=CustomResource

# In a specific namespace
k8sgpt analyze --explain --filter=CustomResource --namespace=production
```

### Example Output

```
AI Provider: openai

0: CustomResource/Certificate(default/example-cert)
- Error: Condition Ready is False (reason: Failed): Certificate issuance failed
- Details: The certificate 'example-cert' in namespace 'default' failed to issue.
  The Let's Encrypt challenge validation failed due to DNS propagation issues.
  Recommendation: Check DNS records and retry certificate issuance.

1: CustomResource/Application(argocd/my-app)
- Error: Health status is Degraded
- Details: The ArgoCD application 'my-app' is in a Degraded state.
  This typically indicates that deployed resources are not healthy.
  Recommendation: Check application logs and pod status.
```

## Best Practices

### 1. Start with Generic Checks
Begin with just `enabled: true` to see what issues are detected across all CRDs.

### 2. Add Specific Configurations Gradually
Add custom configurations for critical CRDs that need specialized health checks.

### 3. Use Exclusions to Reduce Noise
Exclude CRDs that generate false positives or are less critical.

### 4. Combine with Other Analyzers
Use the CRD analyzer alongside built-in analyzers for comprehensive cluster observability.

### 5. Monitor Performance
If you have many CRDs, the analysis may take longer. Use exclusions to optimize.

## Troubleshooting

### Analyzer Not Running
- Verify `enabled: true` is set in configuration
- Check that `CustomResource` is in active filters: `k8sgpt filters list`
- Ensure configuration file is in the correct location

### No Issues Detected
- Verify CRDs are actually installed: `kubectl get crds`
- Check if custom resources exist: `kubectl get <crd-name> --all-namespaces`
- Review generic health check patterns - your CRDs may use different status fields

### Too Many False Positives
- Add specific configurations for problematic CRDs in the `include` section
- Use the `exclude` list to skip noisy CRDs
- Review the status patterns your CRDs use and configure accordingly

### Configuration Not Applied
- Restart K8sGPT after configuration changes
- Verify YAML syntax is correct
- Check K8sGPT logs for configuration parsing errors
