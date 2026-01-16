# Development Guide

This comprehensive guide covers all aspects of developing, testing, and releasing the Cluster Assessment Operator.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Development Environment Setup](#development-environment-setup)
- [Working with OLM](#working-with-olm)
- [Adding New Features](#adding-new-features)
- [Testing Workflows](#testing-workflows)
- [Building Components](#building-components)
- [Release Process](#release-process)
- [Debugging](#debugging)

---

## Prerequisites

### Required Tools

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.25+ | Operator development |
| Podman or Docker | Latest | Container builds |
| operator-sdk | v1.42.0+ | OLM bundle/scorecard |
| opm | v1.36.0+ | FBC catalog management |
| golangci-lint | Latest | Code linting |
| oc/kubectl | Latest | Cluster interaction |

### Optional Tools

| Tool | Purpose |
|------|---------|
| Node.js 18+ | Console plugin development |
| Yarn | Console plugin package management |

### Cluster Access

You need access to an OpenShift 4.12+ cluster for integration testing:

```bash
# Verify cluster access
oc whoami
oc get clusterversion
```

---

## Development Environment Setup

### 1. Clone the Repository

```bash
git clone https://github.com/diegobskt/cluster-assessment-operator.git
cd cluster-assessment-operator
```

### 2. Install Dependencies

```bash
# Go dependencies
make deps

# Verify
go mod verify
```

### 3. Verify Setup

```bash
# Build binary
make build

# Run tests
make test

# Run linter
make lint
```

### 4. Run Locally (Against Remote Cluster)

```bash
export KUBECONFIG=~/.kube/config

# Install CRDs first
make install

# Run operator locally
make run
```

---

## Working with OLM

### Understanding the OLM Structure

The operator uses **File-Based Catalog (FBC)** format for OLM integration:

```
bundle/                          # OLM bundle
├── manifests/
│   ├── cluster-assessment-operator.clusterserviceversion.yaml
│   └── assessment.openshift.io_clusterassessments.yaml
├── metadata/
│   └── annotations.yaml
└── tests/
    └── scorecard/

catalogs/                        # FBC catalogs (per OCP version)
├── v4.12/cluster-assessment-operator/catalog.yaml
├── v4.13/cluster-assessment-operator/catalog.yaml
...
└── v4.20/cluster-assessment-operator/catalog.yaml

catalog-templates/               # Templates for catalog generation
├── v4.12.yaml
├── v4.13.yaml
...
└── v4.20.yaml
```

### OLM Channels

| Channel | Purpose | Target Users |
|---------|---------|--------------|
| `stable-v1` | Production-ready releases | Most users |
| `candidate-v1` | Pre-release testing | Testing teams |
| `fast-v1` | Latest features | Early adopters |

### Bundle Validation

```bash
# Validate bundle structure
operator-sdk bundle validate ./bundle --select-optional suite=operatorframework

# Run scorecard tests
make scorecard
```

### Catalog Validation

```bash
# Validate all FBC catalogs
make catalog-validate

# Validate single catalog
opm validate catalogs/v4.14
```

### Deploy via OLM (Quick Testing)

```bash
# Build and push bundle
make bundle-buildx

# Deploy using operator-sdk
make deploy-olm

# Verify
oc get csv -n cluster-assessment-operator

# Clean up
make cleanup-olm
```

### Deploy via CatalogSource (Production)

```bash
# 1. Build and push catalog
make catalog-build-single OCP_VERSION=v4.14
podman push ghcr.io/diegobskt/cluster-assessment-operator-catalog:v4.14

# 2. Create CatalogSource
oc apply -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: cluster-assessment-catalog
  namespace: openshift-marketplace
spec:
  sourceType: grpc
  image: ghcr.io/diegobskt/cluster-assessment-operator-catalog:v4.14
  displayName: Cluster Assessment Operator
  publisher: Community
EOF

# 3. Create Subscription
oc apply -f - <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: cluster-assessment-operator
---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: cluster-assessment-operator
  namespace: cluster-assessment-operator
spec: {}
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: cluster-assessment-operator
  namespace: cluster-assessment-operator
spec:
  channel: stable-v1
  name: cluster-assessment-operator
  source: cluster-assessment-catalog
  sourceNamespace: openshift-marketplace
EOF

# 4. Verify installation
oc get csv -n cluster-assessment-operator
```

---

## Adding New Features

### Adding a New Validator

1. **Create validator package**:

```bash
mkdir -p pkg/validators/myvalidator
```

2. **Implement the Validator interface**:

```go
// pkg/validators/myvalidator/myvalidator.go
package myvalidator

import (
    "context"

    assessmentv1alpha1 "github.com/diegobskt/cluster-assessment-operator/api/v1alpha1"
    "github.com/diegobskt/cluster-assessment-operator/pkg/profiles"
    "github.com/diegobskt/cluster-assessment-operator/pkg/validator"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
    _ = validator.Register(&MyValidator{})
}

type MyValidator struct{}

func (v *MyValidator) Name() string        { return "myvalidator" }
func (v *MyValidator) Description() string { return "Validates XYZ configuration" }
func (v *MyValidator) Category() string    { return "Platform" }

func (v *MyValidator) Validate(ctx context.Context, c client.Client, profile profiles.Profile) ([]assessmentv1alpha1.Finding, error) {
    var findings []assessmentv1alpha1.Finding

    // Query Kubernetes API (read-only)
    // Build findings based on results

    finding := assessmentv1alpha1.Finding{
        ID:             "myval-001",
        Validator:      v.Name(),
        Category:       v.Category(),
        Status:         "PASS", // PASS, WARN, FAIL, INFO
        Title:          "Check passed",
        Description:    "Detailed description",
        Recommendation: "What to do if this fails",
    }
    findings = append(findings, finding)

    return findings, nil
}
```

3. **Import in main.go**:

```go
// main.go
import (
    // ... other imports
    _ "github.com/diegobskt/cluster-assessment-operator/pkg/validators/myvalidator"
)
```

4. **Add tests**:

```go
// pkg/validators/myvalidator/myvalidator_test.go
package myvalidator

import (
    "context"
    "testing"

    "github.com/diegobskt/cluster-assessment-operator/pkg/profiles"
    "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestMyValidator_Validate(t *testing.T) {
    v := &MyValidator{}
    client := fake.NewClientBuilder().Build()
    profile := profiles.GetProfile("production")

    findings, err := v.Validate(context.Background(), client, profile)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if len(findings) == 0 {
        t.Error("expected at least one finding")
    }
}
```

5. **Update documentation**:
   - Add to validators table in README.md
   - Update architecture.md mindmap

### Adding RBAC Permissions

If your validator needs to read new Kubernetes resources:

1. **Update config/rbac/role.yaml**:

```yaml
# Add a new rule
- apiGroups:
    - your.api.group
  resources:
    - yourresources
  verbs:
    - get
    - list
    - watch
```

2. **Regenerate bundle**:

```bash
make bundle
```

### Console Plugin Changes

1. **Navigate to console-plugin directory**:

```bash
cd console-plugin
```

2. **Install dependencies**:

```bash
yarn install
```

3. **Make changes** in `src/components/`

4. **Test locally**:

```bash
yarn lint
yarn test
yarn build
```

5. **Build and push**:

```bash
podman build -t ghcr.io/diegobskt/cluster-assessment-operator-console:v1.2.10 .
podman push ghcr.io/diegobskt/cluster-assessment-operator-console:v1.2.10
```

---

## Testing Workflows

### Unit Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Test specific package
go test -v ./pkg/validators/version/...

# Test with race detection
go test -race ./...
```

### Linting

```bash
# Run linter
make lint

# Format code
make fmt

# Run go vet
make vet
```

### Integration Tests

```bash
# Deploy operator
make deploy

# Run assessment
oc apply -f examples/quick-html-assessment.yaml

# Watch progress
oc get clusterassessment quick-html-assessment -w

# Check status
oc describe clusterassessment quick-html-assessment

# View report
oc get configmap quick-html-assessment-report -n cluster-assessment-operator \
  -o jsonpath='{.data.report\.json}' | jq .
```

### OLM Testing

```bash
# Validate bundle
operator-sdk bundle validate ./bundle

# Run scorecard
make scorecard

# Test OLM installation
make deploy-olm
oc get csv -n cluster-assessment-operator
make cleanup-olm
```

### Pre-Push Verification

Run this before pushing any changes:

```bash
# Full verification
make fmt && make vet && make lint && make test && make build

# If bundle changed
operator-sdk bundle validate ./bundle

# If catalogs changed
make catalog-validate
```

---

## Building Components

### All-in-One Commands

| Command | Description |
|---------|-------------|
| `make build` | Build Go binary |
| `make podman-buildx` | Build + push multi-arch operator |
| `make bundle-buildx` | Build + push multi-arch bundle |
| `make catalog-build` | Build all catalog images |

### Operator Image

```bash
# Multi-arch (amd64 + arm64)
make podman-buildx

# Single arch (amd64 for OpenShift)
make podman-build
make podman-push

# Local architecture
make podman-build-local
```

### Bundle Image

```bash
# Multi-arch
make bundle-buildx

# Single arch
make bundle-build
make bundle-push
```

### Catalog Images

```bash
# All versions
make catalog-build
make catalog-push

# Single version
make catalog-build-single OCP_VERSION=v4.14
podman push ghcr.io/diegobskt/cluster-assessment-operator-catalog:v4.14
```

### Console Plugin Image

```bash
cd console-plugin
podman build -t ghcr.io/diegobskt/cluster-assessment-operator-console:v1.2.10 .
podman push ghcr.io/diegobskt/cluster-assessment-operator-console:v1.2.10
```

---

## Release Process

### Version Update Checklist

When preparing a release, update version in these files:

1. **Makefile**:
   ```makefile
   IMG ?= ghcr.io/diegobskt/cluster-assessment-operator:vX.Y.Z
   BUNDLE_IMG ?= ghcr.io/diegobskt/cluster-assessment-operator-bundle:vX.Y.Z
   ```

2. **bundle/manifests/cluster-assessment-operator.clusterserviceversion.yaml**:
   - `metadata.name`
   - `spec.version`
   - All image references in `spec.install.spec.deployments`
   - Container annotations

3. **All catalog files** (`catalogs/v4.*/cluster-assessment-operator/catalog.yaml`):
   - Bundle image references

4. **CHANGELOG.md**:
   - Add new version section

5. **config/console-plugin/deployment.yaml** (if console plugin changed):
   - Image tag

### Creating a Release

```bash
# 1. Ensure all version updates are committed
git status

# 2. Create and push tag
git tag v1.2.10
git push origin v1.2.10
```

This triggers GitHub Actions to:
- Build multi-arch operator image
- Build multi-arch bundle image
- Build catalog images for all OCP versions (v4.12-v4.20)
- Create GitHub Release with install.yaml

### Pre-Release (RC/Beta/Alpha)

For pre-releases, use appropriate tag suffix:

```bash
git tag v1.3.0-rc.1
git push origin v1.3.0-rc.1
```

Pre-releases are marked as such in GitHub and go to `candidate-v1` and `fast-v1` channels.

---

## Debugging

### Operator Logs

```bash
# View logs
oc logs -n cluster-assessment-operator deploy/cluster-assessment-operator

# Follow logs
oc logs -n cluster-assessment-operator deploy/cluster-assessment-operator -f

# All containers
oc logs -n cluster-assessment-operator deploy/cluster-assessment-operator --all-containers
```

### Assessment Status

```bash
# Get status
oc get clusterassessment <name> -o yaml

# Describe (includes events)
oc describe clusterassessment <name>

# Get summary
oc get clusterassessment <name> -o jsonpath='{.status.summary}'
```

### Report Retrieval

```bash
# List reports
oc get configmaps -n cluster-assessment-operator | grep report

# Get JSON report
oc get configmap <name>-report -n cluster-assessment-operator \
  -o jsonpath='{.data.report\.json}' > report.json

# Get HTML report
oc get configmap <name>-report -n cluster-assessment-operator \
  -o jsonpath='{.data.report\.html}' > report.html
```

### Console Plugin Debugging

```bash
# Check plugin pod
oc get pods -n cluster-assessment-operator | grep plugin

# View plugin logs
oc logs -n cluster-assessment-operator deploy/cluster-assessment-plugin

# Check ConsolePlugin registration
oc get consoleplugin cluster-assessment-plugin -o yaml
```

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Assessment stuck in Pending | Operator not running | Check operator pod status |
| No findings | All validators filtered | Check `spec.validators` and `spec.minSeverity` |
| ConfigMap not created | Storage not enabled | Set `reportStorage.configMap.enabled: true` |
| Console plugin not visible | Plugin not registered | Check ConsolePlugin CR |
| Catalog not visible in OperatorHub | Image pull issues | Verify catalog image is accessible |

---

## Quick Reference

### Make Targets

```bash
# Development
make build          # Build binary
make test           # Run tests
make lint           # Run linter
make run            # Run locally

# Deployment
make deploy         # Deploy to cluster
make undeploy       # Remove from cluster
make deploy-olm     # Deploy via OLM
make cleanup-olm    # Remove OLM deployment

# Images
make podman-buildx  # Build + push operator (multi-arch)
make bundle-buildx  # Build + push bundle (multi-arch)
make catalog-build  # Build all catalogs
make catalog-push   # Push all catalogs

# Validation
make scorecard          # Run OLM scorecard
make catalog-validate   # Validate FBC catalogs
make preflight          # Red Hat certification checks
```

### Useful oc Commands

```bash
# Assessment operations
oc get clusterassessment
oc describe clusterassessment <name>
oc delete clusterassessment <name>

# Operator status
oc get pods -n cluster-assessment-operator
oc logs deploy/cluster-assessment-operator -n cluster-assessment-operator

# OLM status
oc get csv -n cluster-assessment-operator
oc get subscription -n cluster-assessment-operator
oc get catalogsource -n openshift-marketplace
```
