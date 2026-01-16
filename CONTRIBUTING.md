# Contributing to Cluster Assessment Operator

First off, thank you for considering contributing to the Cluster Assessment Operator! 🎉

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

- **Ensure the bug was not already reported** by searching [Issues](https://github.com/diegobskt/cluster-assessment-operator/issues)
- If you're unable to find an open issue, [open a new one](https://github.com/diegobskt/cluster-assessment-operator/issues/new)
- Include a **clear title and description**, as much relevant information as possible

### Suggesting Enhancements

- Open an issue with the `enhancement` label
- Describe the current behavior and explain the behavior you expected
- Explain why this enhancement would be useful

### Pull Requests

1. Fork the repo and create your branch from `main`
2. Make your changes
3. Ensure tests pass: `make test`
4. Ensure linting passes: `make lint`
5. Update documentation if needed
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.25+
- Podman or Docker
- Access to an OpenShift 4.12+ cluster (for testing)
- operator-sdk v1.42.0+ (for OLM bundle validation)
- opm v1.36.0+ (for FBC catalog management)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/cluster-assessment-operator.git
cd cluster-assessment-operator

# Install dependencies
make deps

# Run tests
make test

# Run linter
make lint

# Build locally
make build

# Run locally (requires KUBECONFIG)
make run
```

### Making Changes

1. **Create a branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes** and add tests

3. **Run the test suite**
   ```bash
   make test
   make lint
   ```

4. **Commit your changes**
   ```bash
   git commit -m "feat: add new validator for XYZ"
   ```

### Commit Message Format

We follow [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only
- `test:` - Adding or updating tests
- `refactor:` - Code change that neither fixes a bug nor adds a feature
- `chore:` - Maintenance tasks

### Adding a New Validator

1. Create a new package under `pkg/validators/yourvalidator/`
2. Implement the `validator.Validator` interface
3. Register in `init()` function
4. Import in `main.go`
5. Add tests under `pkg/validators/yourvalidator/yourvalidator_test.go`
6. Update the README validators table

Example structure:
```go
package yourvalidator

import (
    "context"
    assessmentv1alpha1 "github.com/yourorg/cluster-assessment-operator/api/v1alpha1"
    "github.com/yourorg/cluster-assessment-operator/pkg/profiles"
    "github.com/yourorg/cluster-assessment-operator/pkg/validator"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
    _ = validator.Register(&YourValidator{})
}

type YourValidator struct{}

func (v *YourValidator) Name() string     { return "yourvalidator" }
func (v *YourValidator) Category() string { return "YourCategory" }

func (v *YourValidator) Validate(ctx context.Context, c client.Client, profile *profiles.Profile) ([]assessmentv1alpha1.Finding, error) {
    // Implementation
    return nil, nil
}
```

## Testing

### Running Unit Tests

```bash
# All tests
make test

# With coverage report
make test-coverage

# Specific package
go test -v ./pkg/validators/yourvalidator/...

# Run with verbose output
go test -v ./... -count=1
```

### OLM Bundle Validation

Before pushing, always validate your OLM bundle:

```bash
# Validate bundle structure
operator-sdk bundle validate ./bundle --select-optional suite=operatorframework

# Run scorecard tests (requires cluster access)
make scorecard

# Validate FBC catalogs
make catalog-validate
```

### Integration Testing

```bash
# Deploy operator directly (without OLM)
make deploy

# Run a quick assessment
oc apply -f examples/quick-html-assessment.yaml
oc get clusterassessment quick-html-assessment -w

# Check findings
oc get clusterassessment quick-html-assessment -o jsonpath='{.status.summary}'

# View the report
oc get configmap quick-html-assessment-report -n cluster-assessment-operator \
  -o jsonpath='{.data.report\.html}' > report.html
```

### Testing with OLM

```bash
# Build and push bundle image
make bundle-buildx

# Deploy via OLM
make deploy-olm

# Verify installation
oc get csv -n cluster-assessment-operator

# Clean up
make cleanup-olm
```

### Pre-Push Checklist

Before pushing changes, run this checklist:

```bash
# 1. Format and lint
make fmt
make vet
make lint

# 2. Run tests
make test

# 3. Build binary
make build

# 4. Validate bundle (if changed)
operator-sdk bundle validate ./bundle

# 5. Validate catalogs (if changed)
make catalog-validate
```

## Release Process

Releases are automated via GitHub Actions. To create a release:

1. Update `CHANGELOG.md` with the new version and changes
2. Update version references in these files:
   - `Makefile` (IMG and BUNDLE_IMG variables)
   - `bundle/manifests/cluster-assessment-operator.clusterserviceversion.yaml`
   - All catalog files in `catalogs/v4.*/cluster-assessment-operator/catalog.yaml`
3. Commit and push changes
4. Create and push a version tag:
   ```bash
   git tag v1.x.x
   git push origin v1.x.x
   ```

This triggers:
- Multi-arch operator + bundle image builds (amd64 + arm64)
- FBC catalog images for OCP v4.12-v4.20
- GitHub Release with install.yaml
- Auto-generated PR to update FBC catalogs

### FBC Catalog Validation

Before release, validate catalogs locally:
```bash
make catalog-validate
operator-sdk bundle validate ./bundle --select-optional suite=operatorframework
```

---

## Pushing Individual Components

You can build and push individual components without doing a full release:

### Operator Image Only

```bash
# Build and push operator image
make podman-buildx

# Or for single architecture
make podman-build
make podman-push
```

### Bundle Image Only

```bash
# Build and push bundle
make bundle-buildx

# Or for single architecture
make bundle-build
make bundle-push
```

### Catalog Images Only

```bash
# Build and push all catalog versions
make catalog-build
make catalog-push

# Or build a single version
make catalog-build-single OCP_VERSION=v4.14
podman push ghcr.io/diegobskt/cluster-assessment-operator-catalog:v4.14
```

### Console Plugin Only

```bash
cd console-plugin

# Build the plugin
yarn install
yarn build

# Build container image
podman build -t ghcr.io/diegobskt/cluster-assessment-operator-console:v1.2.10 .
podman push ghcr.io/diegobskt/cluster-assessment-operator-console:v1.2.10

# Deploy to cluster
oc apply -f ../config/console-plugin/
```

---

## Console Plugin Development

The console plugin is a React TypeScript application that integrates with the OpenShift console.

### Setup

```bash
cd console-plugin

# Install dependencies
yarn install

# Start development server (requires bridge to OpenShift)
yarn start
```

### Building

```bash
# Production build
yarn build

# Lint code
yarn lint

# Run tests
yarn test
```

### Plugin Architecture

- **Source**: `console-plugin/src/`
- **Components**: `src/components/` - React components
- **Extensions**: `console-extensions.json` - Console extension points
- **Styles**: `src/components/styles.css`

### Key Files

| File | Purpose |
|------|---------|
| `package.json` | Dependencies and plugin metadata |
| `console-extensions.json` | OpenShift console integration points |
| `webpack.config.ts` | Build configuration |
| `nginx.conf` | Production server configuration |

---

## Project Structure Reference

```
cluster-assessment-operator/
├── api/v1alpha1/              # CRD types (ClusterAssessment)
├── bundle/                    # OLM bundle manifests
├── catalogs/                  # FBC catalogs per OCP version
├── catalog-templates/         # FBC generation templates
├── config/
│   ├── crd/                   # CRD YAML definitions
│   ├── rbac/                  # ClusterRole and bindings
│   ├── manager/               # Operator deployment
│   ├── console-plugin/        # Console plugin deployment
│   └── samples/               # Example CRs
├── console-plugin/            # React TypeScript console UI
├── controllers/               # Reconciliation logic
├── docs/                      # Additional documentation
├── examples/                  # Sample assessments
├── pkg/
│   ├── metrics/               # Prometheus metrics
│   ├── profiles/              # Production/Development profiles
│   ├── report/                # Report generation (JSON/HTML/PDF)
│   ├── validator/             # Validator registry
│   └── validators/            # 18 assessment validators
├── Dockerfile                 # Operator container
├── bundle.Dockerfile          # OLM bundle container
├── catalog.Dockerfile         # FBC catalog container
├── Makefile                   # Build targets
└── main.go                    # Operator entrypoint
```

---

## Questions?

Feel free to open an issue with the `question` label.
