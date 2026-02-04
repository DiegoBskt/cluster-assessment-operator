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

// AssessmentProfileSpec defines a custom assessment profile with overridable thresholds.
type AssessmentProfileSpec struct {
	// Description explains the profile's purpose.
	// +optional
	Description string `json:"description,omitempty"`

	// BasedOn specifies the built-in profile to inherit defaults from.
	// Valid values: "production", "development". Defaults to "production".
	// +kubebuilder:validation:Enum=production;development
	// +kubebuilder:default=production
	// +optional
	BasedOn string `json:"basedOn,omitempty"`

	// Thresholds overrides specific threshold values from the base profile.
	// Only fields that are set will override the base; unset fields inherit defaults.
	// +optional
	Thresholds *ThresholdOverrides `json:"thresholds,omitempty"`

	// EnabledValidators lists validators to enable. If set, only these validators run.
	// Takes precedence over DisabledValidators.
	// +optional
	EnabledValidators []string `json:"enabledValidators,omitempty"`

	// DisabledValidators lists validators to skip. Ignored if EnabledValidators is set.
	// +optional
	DisabledValidators []string `json:"disabledValidators,omitempty"`

	// DisabledChecks lists specific check IDs to skip across all validators.
	// +optional
	DisabledChecks []string `json:"disabledChecks,omitempty"`
}

// ThresholdOverrides allows overriding individual threshold values from the base profile.
// All fields are pointers: nil means "inherit from base profile".
type ThresholdOverrides struct {
	// MinControlPlaneNodes is the minimum expected control plane nodes.
	// +optional
	MinControlPlaneNodes *int `json:"minControlPlaneNodes,omitempty"`

	// MinWorkerNodes is the minimum expected worker nodes.
	// +optional
	MinWorkerNodes *int `json:"minWorkerNodes,omitempty"`

	// MaxPodsPerNode is the maximum recommended pods per node.
	// +optional
	MaxPodsPerNode *int `json:"maxPodsPerNode,omitempty"`

	// MaxClusterAdminBindings is the maximum acceptable cluster-admin bindings.
	// +optional
	MaxClusterAdminBindings *int `json:"maxClusterAdminBindings,omitempty"`

	// RequireNetworkPolicy requires NetworkPolicy in namespaces.
	// +optional
	RequireNetworkPolicy *bool `json:"requireNetworkPolicy,omitempty"`

	// RequireResourceQuotas requires ResourceQuotas in namespaces.
	// +optional
	RequireResourceQuotas *bool `json:"requireResourceQuotas,omitempty"`

	// RequireLimitRanges requires LimitRanges in namespaces.
	// +optional
	RequireLimitRanges *bool `json:"requireLimitRanges,omitempty"`

	// MaxDaysWithoutUpdate is the maximum days since the last cluster update.
	// +optional
	MaxDaysWithoutUpdate *int `json:"maxDaysWithoutUpdate,omitempty"`

	// AllowPrivilegedContainers determines if privileged containers trigger warnings.
	// +optional
	AllowPrivilegedContainers *bool `json:"allowPrivilegedContainers,omitempty"`

	// RequireDefaultStorageClass requires a default StorageClass to be configured.
	// +optional
	RequireDefaultStorageClass *bool `json:"requireDefaultStorageClass,omitempty"`
}

// AssessmentProfileStatus defines the observed state of an AssessmentProfile.
type AssessmentProfileStatus struct {
	// Ready indicates whether the profile has been validated and is usable.
	// +optional
	Ready bool `json:"ready,omitempty"`

	// Message provides details about validation results or errors.
	// +optional
	Message string `json:"message,omitempty"`

	// ResolvedValidatorCount is the number of validators that will run with this profile.
	// +optional
	ResolvedValidatorCount int `json:"resolvedValidatorCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=ap
// +kubebuilder:printcolumn:name="BasedOn",type=string,JSONPath=`.spec.basedOn`
// +kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
// +kubebuilder:printcolumn:name="Validators",type=integer,JSONPath=`.status.resolvedValidatorCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AssessmentProfile defines a custom assessment profile that can be referenced
// by ClusterAssessment resources. It inherits from a built-in base profile
// and allows overriding specific thresholds and validator selections.
type AssessmentProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AssessmentProfileSpec   `json:"spec,omitempty"`
	Status AssessmentProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AssessmentProfileList contains a list of AssessmentProfile
type AssessmentProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AssessmentProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AssessmentProfile{}, &AssessmentProfileList{})
}
