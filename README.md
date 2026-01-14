# OpenShift Cluster Assessment Operator

A Kubernetes Operator for Red Hat OpenShift that performs read-only assessments of cluster configuration and generates human-readable reports highlighting configuration gaps, unsupported settings, and improvement opportunities.

## Overview

The Cluster Assessment Operator is designed for consulting engagements where customers need visibility into their OpenShift cluster's configuration health. It provides:

- **Read-only assessments**: No automatic remediation or configuration changes
- **Comprehensive validation**: Checks across version, nodes, security, networking, storage, and more
- **Baseline profiles**: Production and development profiles with appropriate thresholds
- **Flexible scheduling**: On-demand or cron-scheduled assessments
- **Multiple report formats**: JSON, HTML, and PDF output stored in ConfigMaps

## Features

| Category | Validations |
|----------|-------------|
| **Platform** | OpenShift version, upgrade channel, lifecycle status, MachineConfig health, API server, etcd |
| **Infrastructure** | Node count, conditions, roles, OS consistency, resource allocation |
| **Security** | Cluster-admin bindings, privileged pods, RBAC patterns, etcd encryption, audit logging |
| **Networking** | CNI type, NetworkPolicies, ingress configuration |
| **Storage** | StorageClasses, default SC, CSI driver supportability |
| **Observability** | Monitoring configuration, user workload monitoring, operator status |
| **Compatibility** | Deprecated APIs, missing probes, resource specifications |

## Installation

### Prerequisites

- OpenShift 4.12+
- `oc` CLI with cluster-admin access
- Podman (for building images)

### Quick Install

```bash
# Apply all manifests
oc apply -f config/crd/bases/
oc apply -f config/rbac/
oc apply -f config/manager/
```

### Build from Source

```bash
# Build the operator image
make podman-build IMG=your-registry/cluster-assessment-operator:v1.0.0

# Push to registry
make podman-push IMG=your-registry/cluster-assessment-operator:v1.0.0

# Deploy
make deploy IMG=your-registry/cluster-assessment-operator:v1.0.0
```

## Usage

### One-Time Assessment with Multiple Report Formats

```yaml
apiVersion: assessment.openshift.io/v1alpha1
kind: ClusterAssessment
metadata:
  name: production-assessment
spec:
  profile: production
  reportStorage:
    configMap:
      enabled: true
      format: "json,html,pdf"  # Generate all formats
```

Apply and check results:

```bash
oc apply -f config/samples/assessment_v1alpha1_clusterassessment.yaml

# Watch progress
oc get clusterassessment -w

# View findings in CR status
oc get clusterassessment production-assessment -o jsonpath='{.status.findings}' | jq .
```

### Scheduled Assessment

```yaml
apiVersion: assessment.openshift.io/v1alpha1
kind: ClusterAssessment
metadata:
  name: weekly-assessment
spec:
  schedule: "0 2 * * 0"  # Every Sunday at 2 AM
  profile: production
  reportStorage:
    configMap:
      enabled: true
      format: "json,html"
```

### Development Profile

```yaml
apiVersion: assessment.openshift.io/v1alpha1
kind: ClusterAssessment
metadata:
  name: dev-assessment
spec:
  profile: development  # Relaxed thresholds
  validators:           # Run specific validators only
    - version
    - nodes
    - security
  reportStorage:
    configMap:
      enabled: true
      format: "html"  # HTML only for quick review
```

## Report Formats

The operator supports three report formats:

| Format | Description | Use Case |
|--------|-------------|----------|
| **json** | Machine-readable JSON | Integration with other tools, automated processing |
| **html** | Styled HTML with color-coded findings | Quick browser viewing, sharing via email |
| **pdf** | Professional PDF with visual summaries | Executive reports, documentation, archiving |

Specify multiple formats with comma separation: `format: "json,html,pdf"`

### Accessing Reports

```bash
# List available report files
oc get configmap production-assessment-report -n openshift-cluster-assessment -o jsonpath='{.data}' | jq -r 'keys'

# Extract HTML report
oc get configmap production-assessment-report -n openshift-cluster-assessment \
  -o jsonpath='{.data.report\.html}' > report.html

# Extract PDF report (stored as base64 in binaryData)
oc get configmap production-assessment-report -n openshift-cluster-assessment \
  -o jsonpath='{.binaryData.report\.pdf}' | base64 -d > report.pdf

# Open HTML in browser
open report.html
```

## Baseline Profiles

### Production Profile

Strict enterprise requirements for reliability, security, and supportability:

- Minimum 3 control plane nodes
- Minimum 3 worker nodes
- Network policies required
- Resource quotas required
- No privileged containers allowed
- Updates expected within 90 days

### Development Profile

Relaxed requirements for dev/test environments:

- Single node supported
- Network policies optional
- Privileged containers allowed
- Updates expected within 180 days

## Validators

The operator includes 9 validators:

| Validator | Category | Checks |
|-----------|----------|--------|
| `version` | Platform | OpenShift version, upgrade channel, update availability, lifecycle status |
| `nodes` | Infrastructure | Node count, conditions, roles, OS consistency, resource pressure |
| `machineconfig` | Platform | MachineConfigPool health, custom MachineConfigs |
| `apiserver` | Platform | API server status, etcd health, encryption, audit logging |
| `security` | Security | Cluster-admin bindings, privileged pods, RBAC patterns |
| `networking` | Networking | CNI type, NetworkPolicies, ingress configuration |
| `storage` | Storage | StorageClasses, default SC, CSI drivers |
| `monitoring` | Observability | Monitoring config, user workload monitoring |
| `deprecation` | Compatibility | Deprecated patterns, missing probes, resource limits |

## Report Structure

Reports include cluster info, summary with score, and detailed findings:

```json
{
  "metadata": {
    "generatedAt": "2026-01-14T10:30:00Z",
    "profile": "production"
  },
  "clusterInfo": {
    "clusterVersion": "4.14.8",
    "platform": "AWS",
    "nodeCount": 9
  },
  "summary": {
    "totalChecks": 40,
    "passCount": 28,
    "warnCount": 8,
    "failCount": 3,
    "infoCount": 1,
    "score": 78
  },
  "findings": [
    {
      "id": "security-cluster-admin-excessive",
      "status": "WARN",
      "title": "Excessive Cluster-Admin Bindings",
      "description": "Found 8 non-system cluster-admin bindings",
      "impact": "Increases attack surface",
      "recommendation": "Apply least privilege principle"
    }
  ]
}
```

## Extending the Operator

### Adding New Validators

1. Create a new package under `pkg/validators/`
2. Implement the `Validator` interface:

```go
type Validator interface {
    Name() string
    Description() string
    Category() string
    Validate(ctx context.Context, client client.Client, profile profiles.Profile) ([]Finding, error)
}
```

3. Register in `init()`:

```go
func init() {
    validator.Register(&MyValidator{})
}
```

4. Import in `main.go`:

```go
import _ "github.com/.../pkg/validators/myvalidator"
```

### Adding Profile Thresholds

Edit `pkg/profiles/profiles.go` to add new threshold fields and configure them per profile.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    ClusterAssessment CR                      │
│  (spec: profile, schedule, validators, reportStorage)       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Assessment Controller                      │
│  - Triggers assessments (on-demand or scheduled)            │
│  - Coordinates validators                                   │
│  - Updates CR status                                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Validator Registry                        │
│  - version, nodes, security, networking, storage, ...       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Report Generator                          │
│  - JSON / HTML / PDF output                                 │
│  - CR status / ConfigMap / Git                              │
└─────────────────────────────────────────────────────────────┘
```

## RBAC

The operator uses minimal, read-only permissions for cluster inspection:

- `get`, `list`, `watch` on: nodes, pods, namespaces, configmaps, secrets, etc.
- `get`, `list`, `watch` on: config.openshift.io/*, machineconfiguration.openshift.io/*
- `create`, `update` only on: ConfigMaps (for report storage)

See [config/rbac/role.yaml](config/rbac/role.yaml) for full RBAC configuration.

## Development

```bash
# Run locally
make run

# Run tests
make test

# Build binary
make build

# Build container image
make podman-build

# Push container image
make podman-push
```

## License

Apache License 2.0