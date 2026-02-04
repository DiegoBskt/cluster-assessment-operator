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

// AssessmentSnapshotSpec identifies which assessment produced this snapshot.
type AssessmentSnapshotSpec struct {
	// AssessmentName is the name of the source ClusterAssessment.
	AssessmentName string `json:"assessmentName"`

	// Profile is the profile name used for this run.
	Profile string `json:"profile"`
}

// FindingSnapshot is a compact representation of a finding for historical storage.
// It omits description/impact/recommendation to reduce etcd storage.
type FindingSnapshot struct {
	// ID is the unique identifier for this finding type.
	ID string `json:"id"`

	// Validator is the name of the validator that produced this finding.
	Validator string `json:"validator"`

	// Category groups related findings.
	Category string `json:"category"`

	// Status indicates the finding severity.
	// +kubebuilder:validation:Enum=PASS;WARN;FAIL;INFO
	Status FindingStatus `json:"status"`

	// Title is a short, human-readable title.
	Title string `json:"title"`

	// Resource is the name of the Kubernetes resource involved.
	// +optional
	Resource string `json:"resource,omitempty"`

	// Namespace is the namespace of the resource, if applicable.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// DeltaSummary summarizes changes from the previous assessment snapshot.
type DeltaSummary struct {
	// NewFindings are finding IDs that appeared in this run but not the previous.
	// +optional
	NewFindings []string `json:"newFindings,omitempty"`

	// ResolvedFindings are finding IDs from the previous run that are no longer present.
	// +optional
	ResolvedFindings []string `json:"resolvedFindings,omitempty"`

	// RegressionFindings are findings whose status worsened (e.g., WARN -> FAIL).
	// +optional
	RegressionFindings []string `json:"regressionFindings,omitempty"`

	// ImprovedFindings are findings whose status improved (e.g., FAIL -> WARN or PASS).
	// +optional
	ImprovedFindings []string `json:"improvedFindings,omitempty"`

	// ScoreDelta is the score change from the previous run (positive = improved).
	// +optional
	ScoreDelta *int `json:"scoreDelta,omitempty"`
}

// AssessmentSnapshotStatus holds the snapshot data captured at assessment completion.
type AssessmentSnapshotStatus struct {
	// RunTime is when the assessment completed.
	RunTime metav1.Time `json:"runTime"`

	// Summary is the assessment summary at this point in time.
	Summary AssessmentSummary `json:"summary"`

	// ClusterInfo is the cluster state at this point in time.
	// +optional
	ClusterInfo ClusterInfo `json:"clusterInfo,omitempty"`

	// Findings is the compact list of findings.
	// +optional
	Findings []FindingSnapshot `json:"findings,omitempty"`

	// Delta summarizes changes from the previous snapshot.
	// +optional
	Delta *DeltaSummary `json:"delta,omitempty"`

	// PreviousSnapshotName links to the preceding snapshot for traversal.
	// +optional
	PreviousSnapshotName string `json:"previousSnapshotName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=as
// +kubebuilder:printcolumn:name="Assessment",type=string,JSONPath=`.spec.assessmentName`
// +kubebuilder:printcolumn:name="Score",type=integer,JSONPath=`.status.summary.score`
// +kubebuilder:printcolumn:name="Pass",type=integer,JSONPath=`.status.summary.passCount`
// +kubebuilder:printcolumn:name="Warn",type=integer,JSONPath=`.status.summary.warnCount`
// +kubebuilder:printcolumn:name="Fail",type=integer,JSONPath=`.status.summary.failCount`
// +kubebuilder:printcolumn:name="Run Time",type=date,JSONPath=`.status.runTime`

// AssessmentSnapshot is a point-in-time record of an assessment run,
// used for historical tracking and trend analysis.
type AssessmentSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AssessmentSnapshotSpec   `json:"spec,omitempty"`
	Status AssessmentSnapshotStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AssessmentSnapshotList contains a list of AssessmentSnapshot
type AssessmentSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AssessmentSnapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AssessmentSnapshot{}, &AssessmentSnapshotList{})
}
