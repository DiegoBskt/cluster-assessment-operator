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

package profiles

import (
	"testing"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func intPtr(i int) *int    { return &i }
func boolPtr(b bool) *bool { return &b }

func TestMergeProfile_InheritsBaseDefaults(t *testing.T) {
	custom := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "custom-profile"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn: "production",
		},
	}

	result := mergeProfile(custom)

	if result.Name != "custom-profile" {
		t.Errorf("Expected name 'custom-profile', got %q", result.Name)
	}
	// Should inherit all production defaults
	if result.Thresholds.MinControlPlaneNodes != 3 {
		t.Errorf("Expected MinControlPlaneNodes=3 (production default), got %d", result.Thresholds.MinControlPlaneNodes)
	}
	if result.Thresholds.MinWorkerNodes != 3 {
		t.Errorf("Expected MinWorkerNodes=3 (production default), got %d", result.Thresholds.MinWorkerNodes)
	}
	if result.Thresholds.MaxClusterAdminBindings != 5 {
		t.Errorf("Expected MaxClusterAdminBindings=5 (production default), got %d", result.Thresholds.MaxClusterAdminBindings)
	}
	if !result.Thresholds.RequireNetworkPolicy {
		t.Error("Expected RequireNetworkPolicy=true (production default)")
	}
}

func TestMergeProfile_BasedOnDevelopment(t *testing.T) {
	custom := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-custom"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn: "development",
		},
	}

	result := mergeProfile(custom)

	if result.Thresholds.MinControlPlaneNodes != 1 {
		t.Errorf("Expected MinControlPlaneNodes=1 (development default), got %d", result.Thresholds.MinControlPlaneNodes)
	}
	if result.Thresholds.MaxClusterAdminBindings != 20 {
		t.Errorf("Expected MaxClusterAdminBindings=20 (development default), got %d", result.Thresholds.MaxClusterAdminBindings)
	}
	if result.Thresholds.RequireNetworkPolicy {
		t.Error("Expected RequireNetworkPolicy=false (development default)")
	}
}

func TestMergeProfile_DefaultsToProductionBase(t *testing.T) {
	custom := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "no-base"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			// BasedOn is empty
		},
	}

	result := mergeProfile(custom)

	// Should default to production
	if result.Thresholds.MinControlPlaneNodes != 3 {
		t.Errorf("Expected MinControlPlaneNodes=3 (production default), got %d", result.Thresholds.MinControlPlaneNodes)
	}
}

func TestMergeProfile_OverridesIntThresholds(t *testing.T) {
	custom := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "strict"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn: "production",
			Thresholds: &assessmentv1alpha1.ThresholdOverrides{
				MaxClusterAdminBindings: intPtr(2),
				MaxDaysWithoutUpdate:    intPtr(30),
			},
		},
	}

	result := mergeProfile(custom)

	// Overridden values
	if result.Thresholds.MaxClusterAdminBindings != 2 {
		t.Errorf("Expected MaxClusterAdminBindings=2, got %d", result.Thresholds.MaxClusterAdminBindings)
	}
	if result.Thresholds.MaxDaysWithoutUpdate != 30 {
		t.Errorf("Expected MaxDaysWithoutUpdate=30, got %d", result.Thresholds.MaxDaysWithoutUpdate)
	}
	// Inherited values
	if result.Thresholds.MinControlPlaneNodes != 3 {
		t.Errorf("Expected MinControlPlaneNodes=3 (inherited), got %d", result.Thresholds.MinControlPlaneNodes)
	}
	if result.Thresholds.MinWorkerNodes != 3 {
		t.Errorf("Expected MinWorkerNodes=3 (inherited), got %d", result.Thresholds.MinWorkerNodes)
	}
}

func TestMergeProfile_OverridesBoolThresholds(t *testing.T) {
	custom := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "relaxed-prod"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn: "production",
			Thresholds: &assessmentv1alpha1.ThresholdOverrides{
				RequireNetworkPolicy:  boolPtr(false),
				RequireResourceQuotas: boolPtr(false),
			},
		},
	}

	result := mergeProfile(custom)

	// Overridden values
	if result.Thresholds.RequireNetworkPolicy {
		t.Error("Expected RequireNetworkPolicy=false (overridden)")
	}
	if result.Thresholds.RequireResourceQuotas {
		t.Error("Expected RequireResourceQuotas=false (overridden)")
	}
	// Inherited values
	if !result.Thresholds.RequireLimitRanges {
		t.Error("Expected RequireLimitRanges=true (inherited from production)")
	}
	if result.Thresholds.AllowPrivilegedContainers {
		t.Error("Expected AllowPrivilegedContainers=false (inherited from production)")
	}
}

func TestMergeProfile_AllThresholdsOverridden(t *testing.T) {
	custom := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "full-override"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn: "development",
			Thresholds: &assessmentv1alpha1.ThresholdOverrides{
				MinControlPlaneNodes:       intPtr(5),
				MinWorkerNodes:             intPtr(10),
				MaxPodsPerNode:             intPtr(100),
				MaxClusterAdminBindings:    intPtr(1),
				RequireNetworkPolicy:       boolPtr(true),
				RequireResourceQuotas:      boolPtr(true),
				RequireLimitRanges:         boolPtr(true),
				MaxDaysWithoutUpdate:       intPtr(14),
				AllowPrivilegedContainers:  boolPtr(false),
				RequireDefaultStorageClass: boolPtr(true),
			},
		},
	}

	result := mergeProfile(custom)

	if result.Thresholds.MinControlPlaneNodes != 5 {
		t.Errorf("Expected MinControlPlaneNodes=5, got %d", result.Thresholds.MinControlPlaneNodes)
	}
	if result.Thresholds.MinWorkerNodes != 10 {
		t.Errorf("Expected MinWorkerNodes=10, got %d", result.Thresholds.MinWorkerNodes)
	}
	if result.Thresholds.MaxPodsPerNode != 100 {
		t.Errorf("Expected MaxPodsPerNode=100, got %d", result.Thresholds.MaxPodsPerNode)
	}
	if result.Thresholds.MaxClusterAdminBindings != 1 {
		t.Errorf("Expected MaxClusterAdminBindings=1, got %d", result.Thresholds.MaxClusterAdminBindings)
	}
	if !result.Thresholds.RequireNetworkPolicy {
		t.Error("Expected RequireNetworkPolicy=true")
	}
	if !result.Thresholds.RequireResourceQuotas {
		t.Error("Expected RequireResourceQuotas=true")
	}
	if !result.Thresholds.RequireLimitRanges {
		t.Error("Expected RequireLimitRanges=true")
	}
	if result.Thresholds.MaxDaysWithoutUpdate != 14 {
		t.Errorf("Expected MaxDaysWithoutUpdate=14, got %d", result.Thresholds.MaxDaysWithoutUpdate)
	}
	if result.Thresholds.AllowPrivilegedContainers {
		t.Error("Expected AllowPrivilegedContainers=false")
	}
	if !result.Thresholds.RequireDefaultStorageClass {
		t.Error("Expected RequireDefaultStorageClass=true")
	}
}

func TestMergeProfile_NilThresholds(t *testing.T) {
	custom := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "no-thresholds"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn:    "production",
			Thresholds: nil,
		},
	}

	result := mergeProfile(custom)

	// Should be identical to production
	prod := GetProfile("production")
	if result.Thresholds != prod.Thresholds {
		t.Error("Expected thresholds to match production defaults when no overrides given")
	}
}

func TestMergeProfile_Description(t *testing.T) {
	custom := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "described"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			Description: "Custom profile for finance",
			BasedOn:     "production",
		},
	}

	result := mergeProfile(custom)

	if result.Description != "Custom profile for finance" {
		t.Errorf("Expected description 'Custom profile for finance', got %q", result.Description)
	}
}

func TestMergeProfile_EnabledValidators(t *testing.T) {
	custom := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "security-only"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn:           "production",
			EnabledValidators: []string{"security", "compliance"},
		},
	}

	result := mergeProfile(custom)

	if len(result.EnabledValidators) != 2 {
		t.Errorf("Expected 2 enabled validators, got %d", len(result.EnabledValidators))
	}
	if result.EnabledValidators[0] != "security" || result.EnabledValidators[1] != "compliance" {
		t.Errorf("Expected [security, compliance], got %v", result.EnabledValidators)
	}
}

func TestMergeProfile_DisabledChecks(t *testing.T) {
	custom := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "skip-checks"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn:        "production",
			DisabledChecks: []string{"costoptimization-idle-nodes", "nodes-os-consistency"},
		},
	}

	result := mergeProfile(custom)

	if len(result.DisabledChecks) != 2 {
		t.Errorf("Expected 2 disabled checks, got %d", len(result.DisabledChecks))
	}
}
