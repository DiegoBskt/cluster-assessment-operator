# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.6] - 2026-02-17

### Fixed
- **PDF Report**: Fixed category bar chart legend fragmenting across 5 pages due to gofpdf auto-pagination mismatch between `CellFormat` and `Rect` calls

## [1.3.5] - 2026-02-17

### Added
- **PrometheusRule Alerting**: 5 alert rules targeting existing operator metrics
  - `ClusterAssessmentScoreLow` (critical, score < 60)
  - `ClusterAssessmentScoreDegraded` (warning, score < 80)
  - `ClusterAssessmentRegressions` (warning, regressions detected)
  - `ClusterAssessmentHighFailCount` (warning, > 5 failed checks)
  - `ClusterAssessmentStale` (warning, no run in 24h)
- **5 New Validators** (total now 23):
  - `podsecurityadmission` - PSA labels, privileged enforcement, restricted enforcement
  - `ingresstls` - Route and Ingress TLS configuration validation
  - `clusterautoscaler` - ClusterAutoscaler and MachineAutoscaler presence and config
  - `oadpbackup` - Velero backup schedules, stale/failed backup detection
  - `rbacaudit` - Namespace-scoped RBAC audit (cluster-admin bindings, escalation verbs, sensitive resources)
- **Finding Suppression**:
  - New `SuppressionRule` struct with `findingID`, `reason`, optional `expiresAt`
  - `spec.suppressions[]` to configure suppressed findings per assessment
  - Suppressed findings remain visible but excluded from score calculation
  - Automatic expiration support for temporary suppressions

### Fixed
- **Console Plugin**: Added missing PDF report format option in Create Assessment modal

### Changed
- RBAC role updated with permissions for `route.openshift.io`, `autoscaling.openshift.io`, `machine.openshift.io`, `velero.io`, `oadp.openshift.io`, `monitoring.coreos.com/prometheusrules`
- CRD updated with `suppressions`, `suppressed`, and `suppressionReason` fields

## [1.3.1] - 2026-02-04

### Fixed
- **TypeScript Types**: Added missing `delta` and `snapshotCount` fields to `ClusterAssessment` status interface
- **TypeScript Types**: Moved `DeltaSummary` interface to shared types file; `DeltaBanner` now imports from shared types
- **ESLint**: Removed `as any` cast in `AssessmentDetails.tsx` by properly typing the status object
- **Lint**: Fixed unchecked `os.RemoveAll` return value in git export controller (errcheck)
- **Lint**: Added unreachable `return` after `t.Fatal` in snapshot tests to satisfy staticcheck SA5011

## [1.3.0] - 2026-02-04

### Added
- **Custom Assessment Profiles** (new CRD):
  - New `AssessmentProfile` cluster-scoped CRD for defining custom threshold profiles
  - Pointer-based `ThresholdOverrides` with nil-means-inherit semantics from base profiles
  - Profile resolver that checks built-in names first, then looks up `AssessmentProfile` CRs
  - `AssessmentProfileReconciler` validates profiles and sets `status.ready`
  - `ClusterAssessmentSpec.Profile` now accepts custom profile names (enum constraint removed)
  - Validator filtering via `EnabledValidators`, `DisabledValidators`, and `DisabledChecks`
  - Console plugin dynamically fetches custom profiles for the Create Assessment modal
  - Sample CR: `config/samples/assessment_v1alpha1_assessmentprofile.yaml`

- **Historical Tracking & Trend Analysis** (new CRD):
  - New `AssessmentSnapshot` cluster-scoped CRD for point-in-time history storage
  - `FindingSnapshot` compact format for storage-efficient history (~15KB per snapshot)
  - `DeltaSummary` with new/resolved/regression/improved finding tracking and score delta
  - `SnapshotManager` with `CreateSnapshot`, `ComputeDelta`, and `PruneHistory` methods
  - Configurable retention via `ClusterAssessmentSpec.HistoryLimit` (default 90 snapshots)
  - Delta and snapshot count stored in `ClusterAssessmentStatus` for quick access
  - 4 new Prometheus metrics: `cluster_assessment_score_trend`, `cluster_assessment_new_findings_total`, `cluster_assessment_resolved_findings_total`, `cluster_assessment_regressions_total`
  - Console plugin `TrendChart` component showing score history over time
  - Console plugin `DeltaBanner` showing changes between consecutive runs
  - New "History & Trends" tab in the Assessment Details page

- **Guided Remediation**:
  - New `RemediationGuidance` and `RemediationCommand` types on the `Finding` API
  - `RemediationSafety` enum: `safe-apply`, `requires-review`, `destructive`
  - All 18 validators enhanced with structured remediation data (42+ findings with remediation)
  - Remediation commands are contextual `oc` commands with descriptions and confirmation flags
  - Console plugin `RemediationPanel` with safety badges, `ClipboardCopy` commands, prerequisites, and documentation links
  - HTML report remediation section with styled dark command blocks and safety badges
  - PDF report remediation section with safety labels and formatted commands

### Changed
- RBAC role updated with permissions for `assessmentprofiles` and `assessmentsnapshots`
- `main.go` registers `AssessmentProfileReconciler`
- Controller uses profile resolver instead of direct `GetProfile()` call
- Validator runner filters validators based on profile configuration

## [1.2.44] - 2026-02-04

### Removed
- **Re-run Assessment**: Removed the "Re-run Assessment" button and confirmation modal from the Console Plugin
  - Assessments represent a point-in-time cluster snapshot; users should create a new assessment instead of re-running
  - Removed frontend UI (button, modal, state management, rerun handler)
  - Removed backend `RerunAnnotation` constant and `clearRerunAnnotation` controller logic
  - Removed `assessment.openshift.io/rerun` annotation handling from reconciler

## [1.2.35] - 2026-02-02

### Fixed
- **Release Workflow**: Fixed catalog generation to dynamically render catalogs using `opm` after bundle image is pushed
  - Previous releases had stale catalogs because `catalogs/` directory wasn't regenerated
  - Workflow now uses `opm alpha render-template` to generate fresh catalogs from templates
  - This ensures OLM sees the correct version immediately after release

## [1.2.34] - 2026-02-02

### Fixed
- **Version References**: Fixed bundle CSV and catalog templates to reference correct version
  - v1.2.33 was released with bundle still referencing v1.2.32 images
  - All version references now properly synchronized

## [1.2.33] - 2026-02-02

### Fixed
- **Re-run Assessment Button**: Fixed the "Re-run Assessment" button in the Console Plugin that was not triggering new assessments
  - Changed from status patching to annotation-based trigger mechanism
  - Controller now detects `assessment.openshift.io/rerun` annotation and triggers fresh assessment
  - Previous findings are properly cleared before new assessment runs

## [1.2.32] - 2026-01-29

### Changed
- Updated all catalog files to reference v1.2.32 bundle
- Various bug fixes and stability improvements from v1.2.12-v1.2.31

### Fixed
- **UI Stability**: Prevent UI from going blank by implementing stable assessment data and ensure non-mutating array sorting
- **Security**: Fix stored XSS vulnerability in HTML report generation
- **Security**: Fix cross-tenant secret access in git export functionality
- **Security**: Secure git secret lookup to prevent cross-tenant access
- **Namespace Isolation**: Fix namespace isolation for reports and secrets

### Added
- **Git Export**: Implement git export for assessment reports to remote repositories
- **Empty States**: Add empty states to findings table for better UX
- **Accessibility**: Enhance accessibility for tables and external links
- **Helper Text**: Add helper text for assessment profile selection
- **Report Validation**: Add validation for report formats
- **Disabled States**: Disable Create Assessment modal inputs during submission

### Performance
- Optimize resourcequotas validator with hoisted resource parsing
- Optimize node listing in cluster assessment controller using PartialObjectMetadata
- Optimize etcdbackup validator CronJob listing using PartialObjectMetadata
- Optimize FindingsTable rendering with React.memo
- Hoist `resource.MustParse` out of loops in validators

## [1.2.31] - 2026-01-21

### Added
- **OLM Console Plugin Deployment**: Console plugin is now automatically deployed when operator is installed via OLM
  - Added console plugin deployment to CSV install.spec.deployments
  - Added ConsolePlugin CR and Service to bundle manifests
  - Makefile now syncs console plugin image version in bundle

## [1.2.30] - 2026-01-20

### Changed
- Release preparation and version bump

## [1.2.29] - 2026-01-20

### Fixed
- Update bundle CSV version to v1.2.29 for proper OLM catalog support
- Update all deployment manifests to v1.2.29
- Update catalog templates to reference v1.2.29 bundle

## [1.2.28] - 2026-01-20

### Added
- **VERSION File**: Single source of truth for version management
- **Release Automation**: `scripts/update-catalogs.sh` for automated catalog updates
- **Release Documentation**: `docs/RELEASE.md` with comprehensive release documentation
- **Agent Workflows**: `.agent/workflows` for deploy-operator and cleanup-operator
- **Makefile Targets**: `version`, `update-manifests`, `update-catalogs`, `release-prep`

### Changed
- Improved README.md with better deployment instructions
- Updated all manifests and catalogs to v1.2.28

## [1.2.27] - 2026-01-20

### Fixed
- **RBAC Permissions**: Added permissions for all validator API groups
  - `imageregistry.operator.openshift.io` (configs, imagepruners)
  - `logging.openshift.io` (clusterloggings, clusterlogforwarders)
  - `oadp.openshift.io` (dataprotectionapplications)
  - `operators.coreos.com` (subscriptions, CSVs, installplans, catalogsources)

## [1.2.26] - 2026-01-19

### Fixed
- **Cache Busting**: Aggressive cache-busting for all plugin assets
  - Set no-store on ALL JS, CSS, and JSON files
  - Disabled ETags globally
  - Added `if_modified_since off` to prevent 304 responses
  - Added `Expires: 0` header for maximum compatibility

## [1.2.25] - 2026-01-19

### Changed
- **OpenShift 4.20 Compatibility**: Major upgrade for console plugin
  - Updated `@openshift-console/dynamic-plugin-sdk` to 4.20.0
  - Updated `@openshift-console/dynamic-plugin-sdk-webpack` to 4.20.0
  - Migrated to PatternFly 5.4.14 (tables, forms, modals)
  - Updated table components to use Thead/Tbody/Tr/Th/Td syntax
  - Updated form event handlers to new `(_event, value)` signature
  - Updated CSS variables to `--pf-v5-global-*` format

## [1.2.24] - 2026-01-19

### Fixed
- Aligned package.json with official console plugin template
- Added react-i18next dependency

## [1.2.23] - 2026-01-19

### Fixed
- Use export default function pattern matching official console plugin template

## [1.2.22] - 2026-01-19

### Fixed
- Aligned SDK/webpack versions and exposed modules for module federation compatibility

## [1.2.21] - 2026-01-19

### Fixed
- Updated nginx cache headers to prevent stale cache after plugin updates

## [1.2.20] - 2026-01-19

### Added
- **Theme Support**: Theme-aware styling with PatternFly CSS variables for light/dark mode

## [1.2.19] - 2026-01-19

### Changed
- **Enhanced UI**: Console plugin UI with modern styling and card layout

## [1.2.18] - 2026-01-19

### Fixed
- Restored full console plugin components (Chrome bug was the issue, not the code)

## [1.2.17] - 2026-01-19

### Fixed
- Use explicit `.default` reference in codeRef for lazy loading

## [1.2.16] - 2026-01-19

### Fixed
- Updated SDK to v4.20.0 and webpack to 5.75.0 to fix React error #306

## [1.2.15] - 2026-01-17

### Added
- Minimal plugin for debugging React error #306

## [1.2.14] - 2026-01-16

### Fixed
- Resolved circular imports by creating shared types file

## [1.2.13] - 2026-01-16

### Fixed
- Replaced deprecated Select with FormSelect to fix React error #306

## [1.2.12] - 2026-01-16

### Fixed
- Corrected TypeScript errors in CreateAssessmentModal

## [1.2.11] - 2026-01-16

### Added
- **Create Assessment Modal**: Functional modal for creating new ClusterAssessment resources from the Console UI
  - Profile selection (Production/Development)
  - Report format options (HTML, JSON)

### Changed
- **Improved Console Plugin Styling**: Enhanced visual appearance with card hover effects, better typography, and status highlighting

## [1.2.10] - 2026-01-16

### Fixed
- **Console Plugin TLS**: nginx now serves HTTPS using OpenShift-provided serving certificate
  - Added TLS secret volume mount to deployment
  - Configured nginx with SSL using `/etc/nginx/tls/tls.crt` and `tls.key`
- Console plugin nginx configuration for read-only root filesystem:
  - Changed error/access logs to use `/dev/stderr` and `/dev/stdout` (container best practice)
  - Configured nginx temp paths to use `/tmp` directory
  - Added emptyDir volume mount for `/tmp` in deployment
- Updated all image references and catalogs to v1.2.10

## [1.2.9] - 2026-01-16

### Added
- New CatalogSource using `ghcr.io/diegobskt/cluster-assessment-operator-catalog:v4.20`.

### Changed
- Revised controller logic and documentation across multiple catalog versions.
- Updated Operator to v1.2.9.

## [1.2.8] - 2026-01-16

### Fixed
- Added RBAC permissions for `limitranges`.

## [1.2.7] - 2026-01-16

### Fixed
- Added RBAC permissions for `resourcequotas`.
- Updated namespace references.

## [1.2.6] - 2026-01-16

### Fixed
- Enforced consistent `v`-prefixed image tags for operator releases.

## [1.2.5] - 2026-01-16

### Changed
- Updated FBC catalogs to use `v1.2.4` bundle.

## [1.2.4] - 2026-01-16

### Fixed
- Constrained console plugin build to `amd64` architecture only.

## [1.2.3] - 2026-01-16

### Fixed
- Updated CSV to reference `v1.2.2` operator image correctly.

## [1.2.2] - 2026-01-16

### Changed
- Updated FBC catalogs to `v1.2.0` bundle.

## [1.2.1] - 2026-01-15

### Fixed
- Update bundle CSV to v1.2.0 with correct image references.

## [1.2.0] - 2026-01-15

### Added
- **OpenShift Dynamic Console Plugin**: UI integration for cluster assessments.

### Fixed
- Updated console-plugin to use React 17 and PatternFly 4.

## [1.1.1] - 2026-01-15

### Fixed
- FBC catalog image references now use lowercase (fixes OLM visibility)
- ConfigMap names always include timestamp to prevent overwriting previous reports

## [1.1.0] - 2026-01-15

### Added
- **6 New Validators** (total now 18):
  - `imageregistry` - Registry configuration, storage backend, pruning, replicas
  - `compliance` - Pod Security Admission, OAuth providers, kubeadmin user
  - `resourcequotas` - ResourceQuota coverage, utilization, LimitRanges
  - `logging` - ClusterLogging operator, log forwarding, collector health
  - `costoptimization` - Orphan PVCs, idle deployments, resource specifications
  - `networkpolicyaudit` - Policy coverage, allow-all detection, default deny
- New **Governance** category for resource management validators

### Changed
- Validators are now organized alphabetically in main.go imports

## [1.0.0] - 2026-01-14

### Added
- Initial release of Cluster Assessment Operator
- **12 Validators**:
  - `version` - OpenShift version, upgrade channel, update availability
  - `nodes` - Node count, conditions, roles, OS consistency
  - `machineconfig` - MachineConfigPool health, custom MachineConfigs
  - `apiserver` - API server status, etcd health, encryption, audit logging
  - `operators` - ClusterServiceVersion states, ClusterOperator health
  - `certificates` - TLS certificate expiration, custom certs
  - `etcdbackup` - OADP, Velero, backup CronJob configuration
  - `security` - Cluster-admin bindings, privileged pods, RBAC
  - `networking` - CNI type, NetworkPolicies, ingress configuration
  - `storage` - StorageClasses, default SC, CSI drivers
  - `monitoring` - Cluster monitoring, user workload monitoring
  - `deprecation` - Deprecated patterns, missing probes

- **Report Formats**: JSON, HTML, PDF
- **Report Storage**: ConfigMap with automatic timestamp naming
- **Baseline Profiles**: Production (strict) and Development (relaxed)
- **Scheduled Assessments**: Cron-based scheduling support
- **Severity Filtering**: Filter findings by minimum severity (INFO, PASS, WARN, FAIL)
- **Prometheus Metrics**: Assessment score, findings count, duration
- **OLM Bundle**: Full OLM support with scorecard tests passing
- **Multi-arch Support**: amd64 and arm64 container images
- **Red Hat Certification Ready**:
  - UBI9 base image
  - Required container labels
  - License directory
  - Non-root execution

### Security
- Read-only RBAC (only get, list, watch on cluster resources)
- Non-root container execution (USER 65532)
- Seccomp RuntimeDefault profile
- No privilege escalation

---

## Version History

| Version | Date | Description |
|---------|------|-------------|
| 1.3.6 | 2026-02-17 | PDF category legend fix |
| 1.3.5 | 2026-02-17 | PrometheusRule alerting, 5 new validators, finding suppression, PDF modal fix |
| 1.3.1 | 2026-02-04 | Lint and type fixes |
| 1.3.0 | 2026-02-04 | Custom profiles, historical tracking, guided remediation |
| 1.2.33 | 2026-02-02 | Re-run Assessment button fix |
| 1.2.32 | 2026-01-29 | Security fixes, git export, performance optimizations |
| 1.2.31 | 2026-01-21 | OLM console plugin deployment |
| 1.2.28 | 2026-01-20 | Version management, release automation |
| 1.2.27 | 2026-01-20 | RBAC permissions for all validators |
| 1.2.26 | 2026-01-19 | Aggressive cache busting |
| 1.2.25 | 2026-01-19 | OpenShift 4.20 / PatternFly 5 upgrade |
| 1.2.20 | 2026-01-19 | Theme support (light/dark mode) |
| 1.2.19 | 2026-01-19 | Enhanced UI with card layout |
| 1.2.11 | 2026-01-16 | Create Assessment modal |
| 1.2.10 | 2026-01-16 | Console plugin nginx TLS fix |
| 1.2.0 | 2026-01-15 | OpenShift Dynamic Console Plugin |
| 1.1.0 | 2026-01-15 | 6 new validators (18 total) |
| 1.0.0 | 2026-01-14 | Initial release |

[Unreleased]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.3.6...HEAD
[1.3.6]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.3.5...v1.3.6
[1.3.5]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.3.1...v1.3.5
[1.3.1]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.44...v1.3.0
[1.2.44]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.33...v1.2.44
[1.2.33]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.32...v1.2.33
[1.2.32]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.31...v1.2.32
[1.2.31]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.30...v1.2.31
[1.2.30]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.29...v1.2.30
[1.2.29]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.28...v1.2.29
[1.2.28]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.27...v1.2.28
[1.2.27]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.26...v1.2.27
[1.2.26]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.25...v1.2.26
[1.2.25]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.24...v1.2.25
[1.2.24]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.23...v1.2.24
[1.2.23]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.22...v1.2.23
[1.2.22]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.21...v1.2.22
[1.2.21]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.20...v1.2.21
[1.2.20]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.19...v1.2.20
[1.2.19]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.18...v1.2.19
[1.2.18]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.17...v1.2.18
[1.2.17]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.16...v1.2.17
[1.2.16]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.15...v1.2.16
[1.2.15]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.14...v1.2.15
[1.2.14]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.13...v1.2.14
[1.2.13]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.12...v1.2.13
[1.2.12]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.11...v1.2.12
[1.2.11]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.10...v1.2.11
[1.2.10]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.9...v1.2.10
[1.2.9]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.8...v1.2.9
[1.2.8]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.7...v1.2.8
[1.2.7]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.6...v1.2.7
[1.2.6]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.5...v1.2.6
[1.2.5]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.4...v1.2.5
[1.2.4]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.3...v1.2.4
[1.2.3]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.2...v1.2.3
[1.2.2]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/diegobskt/cluster-assessment-operator/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/diegobskt/cluster-assessment-operator/releases/tag/v1.0.0
