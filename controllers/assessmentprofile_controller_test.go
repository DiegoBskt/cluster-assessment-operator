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

package controllers

import (
	"context"
	"testing"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/profiles"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/validator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// testValidator implements validator.Validator for testing.
type testValidator struct {
	name string
}

func (v *testValidator) Name() string        { return v.name }
func (v *testValidator) Description() string { return "test validator" }
func (v *testValidator) Category() string    { return "Test" }
func (v *testValidator) Validate(_ context.Context, _ client.Client, _ profiles.Profile) ([]assessmentv1alpha1.Finding, error) {
	return nil, nil
}

func newTestRegistry(names ...string) *validator.Registry {
	reg := validator.NewRegistry()
	for _, name := range names {
		_ = reg.Register(&testValidator{name: name})
	}
	return reg
}

func TestValidateProfile_ValidProfile(t *testing.T) {
	reg := newTestRegistry("security", "nodes", "networking")
	r := &AssessmentProfileReconciler{Registry: reg}

	profile := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn: "production",
		},
	}

	ready, message, count := r.validateProfile(profile)

	if !ready {
		t.Errorf("Expected ready=true, got false, message: %s", message)
	}
	if count != 3 {
		t.Errorf("Expected count=3, got %d", count)
	}
}

func TestValidateProfile_InvalidBasedOn(t *testing.T) {
	reg := newTestRegistry()
	r := &AssessmentProfileReconciler{Registry: reg}

	profile := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn: "staging",
		},
	}

	ready, message, _ := r.validateProfile(profile)

	if ready {
		t.Error("Expected ready=false for invalid basedOn")
	}
	if message == "" {
		t.Error("Expected error message for invalid basedOn")
	}
}

func TestValidateProfile_EmptyBasedOnDefaultsToProduction(t *testing.T) {
	reg := newTestRegistry()
	r := &AssessmentProfileReconciler{Registry: reg}

	profile := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec:       assessmentv1alpha1.AssessmentProfileSpec{
			// BasedOn is empty - should default to production and be valid
		},
	}

	ready, message, _ := r.validateProfile(profile)

	if !ready {
		t.Errorf("Expected ready=true for empty basedOn (defaults to production), message: %s", message)
	}
}

func TestValidateProfile_UnknownEnabledValidator(t *testing.T) {
	reg := newTestRegistry("security", "nodes")
	r := &AssessmentProfileReconciler{Registry: reg}

	profile := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn:           "production",
			EnabledValidators: []string{"security", "nonexistent-validator"},
		},
	}

	ready, message, _ := r.validateProfile(profile)

	if ready {
		t.Error("Expected ready=false for unknown enabled validator")
	}
	if message == "" {
		t.Error("Expected error message for unknown validator")
	}
}

func TestValidateProfile_UnknownDisabledValidator(t *testing.T) {
	reg := newTestRegistry("security", "nodes")
	r := &AssessmentProfileReconciler{Registry: reg}

	profile := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn:            "production",
			DisabledValidators: []string{"nonexistent-validator"},
		},
	}

	ready, message, _ := r.validateProfile(profile)

	if ready {
		t.Error("Expected ready=false for unknown disabled validator")
	}
	if message == "" {
		t.Error("Expected error message for unknown validator")
	}
}

func TestValidateProfile_EnabledValidatorsCount(t *testing.T) {
	reg := newTestRegistry("security", "nodes", "networking", "storage", "version")
	r := &AssessmentProfileReconciler{Registry: reg}

	profile := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn:           "production",
			EnabledValidators: []string{"security", "nodes"},
		},
	}

	ready, _, count := r.validateProfile(profile)

	if !ready {
		t.Error("Expected ready=true")
	}
	if count != 2 {
		t.Errorf("Expected count=2 (only enabled validators), got %d", count)
	}
}

func TestValidateProfile_DisabledValidatorsCount(t *testing.T) {
	reg := newTestRegistry("security", "nodes", "networking", "storage", "version")
	r := &AssessmentProfileReconciler{Registry: reg}

	profile := &assessmentv1alpha1.AssessmentProfile{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: assessmentv1alpha1.AssessmentProfileSpec{
			BasedOn:            "production",
			DisabledValidators: []string{"storage", "version"},
		},
	}

	ready, _, count := r.validateProfile(profile)

	if !ready {
		t.Error("Expected ready=true")
	}
	if count != 3 {
		t.Errorf("Expected count=3 (5 total - 2 disabled), got %d", count)
	}
}
