/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterAssessmentSpec defines the desired state of ClusterAssessment
type ClusterAssessmentSpec struct {
	// Schedule in cron format for periodic assessments.
	// Leave empty for one-time assessment triggered on CR creation.
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// Profile specifies the baseline profile to use for assessment.
	// Can be a built-in profile name ("production", "development") or
	// the name of a custom AssessmentProfile CR.
	// +kubebuilder:default=production
	// +optional
	Profile string `json:"profile,omitempty"`

	// Validators is the list of specific validators to run.
	// Leave empty to run all validators.
	// +optional
	Validators []string `json:"validators,omitempty"`

	// Suspend prevents scheduled assessments from running when true.
	// +optional
	Suspend bool `json:"suspend,omitempty"`

	// ReportStorage configures where assessment reports are stored.
	// +optional
	ReportStorage ReportStorageSpec `json:"reportStorage,omitempty"`

	// MinSeverity filters findings to only include this severity level and above.
	// Valid values are: "INFO", "PASS", "WARN", "FAIL"
	// Leave empty to include all findings.
	// +kubebuilder:validation:Enum=INFO;PASS;WARN;FAIL
	// +optional
	MinSeverity string `json:"minSeverity,omitempty"`

	// HistoryLimit is the maximum number of AssessmentSnapshot CRs to retain per assessment.
	// Oldest snapshots are pruned when this limit is exceeded.
	// Set to 0 to disable historical tracking. Defaults to 90.
	// +kubebuilder:default=90
	// +optional
	HistoryLimit *int `json:"historyLimit,omitempty"`

	// Suppressions lists finding IDs to suppress from scoring.
	// Suppressed findings are still collected and visible in reports
	// but marked as suppressed and excluded from score calculation.
	// +optional
	Suppressions []SuppressionRule `json:"suppressions,omitempty"`
}

// ReportStorageSpec configures report storage options
type ReportStorageSpec struct {
	// ConfigMap enables storing the report in a ConfigMap.
	// +optional
	ConfigMap *ConfigMapStorageSpec `json:"configMap,omitempty"`

	// Git enables exporting the report to a Git repository.
	// +optional
	Git *GitStorageSpec `json:"git,omitempty"`
}

// ConfigMapStorageSpec configures ConfigMap storage
type ConfigMapStorageSpec struct {
	// Enabled determines if ConfigMap storage is active.
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// Name is the ConfigMap name. Defaults to <assessment-name>-report.
	// +optional
	Name string `json:"name,omitempty"`

	// Namespace is the namespace where the ConfigMap will be created.
	// Defaults to the operator's namespace if not specified.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Format specifies the report format(s) to generate.
	// Valid values are: "json", "html", "pdf", or combinations like "json,html,pdf"
	// Defaults to "json"
	// +optional
	Format string `json:"format,omitempty"`
}

// GitStorageSpec configures Git repository export
type GitStorageSpec struct {
	// Enabled determines if Git export is active.
	// +optional
	Enabled bool `json:"enabled,omitempty"`

	// URL is the Git repository URL.
	// +optional
	URL string `json:"url,omitempty"`

	// Branch is the target branch. Defaults to "main".
	// +optional
	Branch string `json:"branch,omitempty"`

	// Path is the directory path within the repository.
	// +optional
	Path string `json:"path,omitempty"`

	// SecretRef references a secret containing Git credentials.
	// The secret should contain 'username' and 'password' or 'token' keys.
	// +optional
	SecretRef string `json:"secretRef,omitempty"`

	// SecretNamespace is the namespace of the secret referenced by SecretRef.
	// Required when SecretRef is set, since ClusterAssessment is cluster-scoped.
	// +optional
	SecretNamespace string `json:"secretNamespace,omitempty"`
}

// ClusterAssessmentStatus defines the observed state of ClusterAssessment
type ClusterAssessmentStatus struct {
	// Phase represents the current phase of the assessment.
	// +kubebuilder:validation:Enum=Pending;Running;Completed;Failed
	// +optional
	Phase string `json:"phase,omitempty"`

	// LastRunTime is the timestamp of the last assessment run.
	// +optional
	LastRunTime *metav1.Time `json:"lastRunTime,omitempty"`

	// NextRunTime is the scheduled time for the next assessment (if scheduled).
	// +optional
	NextRunTime *metav1.Time `json:"nextRunTime,omitempty"`

	// ClusterInfo contains metadata about the assessed cluster.
	// +optional
	ClusterInfo ClusterInfo `json:"clusterInfo,omitempty"`

	// Summary provides an overview of assessment results.
	// +optional
	Summary AssessmentSummary `json:"summary,omitempty"`

	// Findings is the list of all assessment findings.
	// +optional
	Findings []Finding `json:"findings,omitempty"`

	// ReportConfigMap is the name of the ConfigMap containing the full report.
	// +optional
	ReportConfigMap string `json:"reportConfigMap,omitempty"`

	// Conditions represent the latest available observations of the assessment's state.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Message provides additional information about the current phase.
	// +optional
	Message string `json:"message,omitempty"`

	// Delta summarizes changes from the previous assessment run.
	// +optional
	Delta *DeltaSummary `json:"delta,omitempty"`

	// SnapshotCount is the number of historical snapshots retained for this assessment.
	// +optional
	SnapshotCount int `json:"snapshotCount,omitempty"`
}

// ClusterInfo contains metadata about the OpenShift cluster
type ClusterInfo struct {
	// ClusterID is the unique identifier of the cluster.
	// +optional
	ClusterID string `json:"clusterID,omitempty"`

	// ClusterVersion is the current OpenShift version.
	// +optional
	ClusterVersion string `json:"clusterVersion,omitempty"`

	// Platform is the infrastructure platform (AWS, Azure, vSphere, etc.).
	// +optional
	Platform string `json:"platform,omitempty"`

	// Channel is the update channel configured for the cluster.
	// +optional
	Channel string `json:"channel,omitempty"`

	// NodeCount is the total number of nodes in the cluster.
	// +optional
	NodeCount int `json:"nodeCount,omitempty"`

	// ControlPlaneNodes is the number of control plane nodes.
	// +optional
	ControlPlaneNodes int `json:"controlPlaneNodes,omitempty"`

	// WorkerNodes is the number of worker nodes.
	// +optional
	WorkerNodes int `json:"workerNodes,omitempty"`
}

// AssessmentSummary provides an overview of assessment results
type AssessmentSummary struct {
	// TotalChecks is the total number of checks performed.
	TotalChecks int `json:"totalChecks"`

	// PassCount is the number of checks that passed.
	PassCount int `json:"passCount"`

	// WarnCount is the number of checks with warnings.
	WarnCount int `json:"warnCount"`

	// FailCount is the number of checks that failed.
	FailCount int `json:"failCount"`

	// InfoCount is the number of informational findings.
	InfoCount int `json:"infoCount"`

	// Score is an optional overall health/maturity score (0-100).
	// +optional
	Score *int `json:"score,omitempty"`

	// ProfileUsed is the baseline profile that was used.
	// +optional
	ProfileUsed string `json:"profileUsed,omitempty"`
}

// Finding represents a single assessment finding
type Finding struct {
	// ID is a unique identifier for this finding type.
	ID string `json:"id"`

	// Validator is the name of the validator that produced this finding.
	Validator string `json:"validator"`

	// Category groups related findings (e.g., "Security", "Networking").
	Category string `json:"category"`

	// Resource is the name of the Kubernetes resource involved.
	// +optional
	Resource string `json:"resource,omitempty"`

	// Namespace is the namespace of the resource, if applicable.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Status indicates the finding severity: PASS, WARN, FAIL, or INFO.
	// +kubebuilder:validation:Enum=PASS;WARN;FAIL;INFO
	Status FindingStatus `json:"status"`

	// Title is a short, human-readable title for the finding.
	Title string `json:"title"`

	// Description explains what was checked and what was found.
	Description string `json:"description"`

	// Impact explains why this finding matters from reliability, security,
	// or supportability perspectives.
	// +optional
	Impact string `json:"impact,omitempty"`

	// Recommendation describes how the configuration could be improved.
	// This is advisory only; no automatic remediation is performed.
	// +optional
	Recommendation string `json:"recommendation,omitempty"`

	// References provides links to relevant documentation.
	// +optional
	References []string `json:"references,omitempty"`

	// Remediation provides structured guidance for resolving this finding.
	// +optional
	Remediation *RemediationGuidance `json:"remediation,omitempty"`

	// Suppressed indicates this finding was matched by a suppression rule
	// and is excluded from score calculation.
	// +optional
	Suppressed bool `json:"suppressed,omitempty"`

	// SuppressionReason explains why this finding was suppressed.
	// +optional
	SuppressionReason string `json:"suppressionReason,omitempty"`
}

// RemediationSafety indicates the safety level of applying the remediation.
// +kubebuilder:validation:Enum="safe-apply";"requires-review";"destructive"
type RemediationSafety string

const (
	// RemediationSafeApply indicates the remediation is safe to apply directly.
	RemediationSafeApply RemediationSafety = "safe-apply"
	// RemediationRequiresReview indicates the remediation should be reviewed before applying.
	RemediationRequiresReview RemediationSafety = "requires-review"
	// RemediationDestructive indicates the remediation may cause service disruption.
	RemediationDestructive RemediationSafety = "destructive"
)

// RemediationGuidance provides structured remediation information for a finding.
type RemediationGuidance struct {
	// Safety indicates the risk level of applying this remediation.
	// +kubebuilder:validation:Enum="safe-apply";"requires-review";"destructive"
	Safety RemediationSafety `json:"safety"`

	// Commands is an ordered list of commands to remediate the finding.
	// +optional
	Commands []RemediationCommand `json:"commands,omitempty"`

	// DocumentationURL links to relevant documentation.
	// +optional
	DocumentationURL string `json:"documentationURL,omitempty"`

	// EstimatedImpact describes what will change when the remediation is applied.
	// +optional
	EstimatedImpact string `json:"estimatedImpact,omitempty"`

	// Prerequisites lists conditions that should be met before applying the remediation.
	// +optional
	Prerequisites []string `json:"prerequisites,omitempty"`
}

// RemediationCommand represents a single command step in a remediation procedure.
type RemediationCommand struct {
	// Command is the shell command to execute.
	Command string `json:"command"`

	// Description explains what this command does.
	// +optional
	Description string `json:"description,omitempty"`

	// RequiresConfirmation indicates this command is potentially dangerous
	// and the user should confirm before executing.
	// +optional
	RequiresConfirmation bool `json:"requiresConfirmation,omitempty"`
}

// SuppressionRule defines a rule for suppressing specific findings.
type SuppressionRule struct {
	// FindingID is the ID of the finding to suppress.
	FindingID string `json:"findingID"`

	// Reason explains why this finding is being suppressed.
	Reason string `json:"reason"`

	// ExpiresAt is an optional expiration time for the suppression.
	// After this time, the finding will no longer be suppressed.
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`
}

// FindingStatus represents the status of a finding
// +kubebuilder:validation:Enum=PASS;WARN;FAIL;INFO
type FindingStatus string

const (
	// FindingStatusPass indicates the check passed with no issues.
	FindingStatusPass FindingStatus = "PASS"
	// FindingStatusWarn indicates a warning that should be reviewed.
	FindingStatusWarn FindingStatus = "WARN"
	// FindingStatusFail indicates a failed check requiring attention.
	FindingStatusFail FindingStatus = "FAIL"
	// FindingStatusInfo indicates informational finding with no action needed.
	FindingStatusInfo FindingStatus = "INFO"
)

// Assessment phase constants
const (
	PhasePending   = "Pending"
	PhaseRunning   = "Running"
	PhaseCompleted = "Completed"
	PhaseFailed    = "Failed"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=ca
// +kubebuilder:printcolumn:name="Profile",type=string,JSONPath=`.spec.profile`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Pass",type=integer,JSONPath=`.status.summary.passCount`
// +kubebuilder:printcolumn:name="Warn",type=integer,JSONPath=`.status.summary.warnCount`
// +kubebuilder:printcolumn:name="Fail",type=integer,JSONPath=`.status.summary.failCount`
// +kubebuilder:printcolumn:name="Last Run",type=date,JSONPath=`.status.lastRunTime`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ClusterAssessment is the Schema for the clusterassessments API.
// It triggers read-only assessments of OpenShift cluster configuration and
// generates human-readable reports with findings and recommendations.
type ClusterAssessment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterAssessmentSpec   `json:"spec,omitempty"`
	Status ClusterAssessmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterAssessmentList contains a list of ClusterAssessment
type ClusterAssessmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterAssessment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterAssessment{}, &ClusterAssessmentList{})
}
