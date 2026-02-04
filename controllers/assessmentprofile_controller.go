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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/profiles"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/validator"
)

// AssessmentProfileReconciler reconciles an AssessmentProfile object
type AssessmentProfileReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Registry *validator.Registry
}

// +kubebuilder:rbac:groups=assessment.openshift.io,resources=assessmentprofiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=assessment.openshift.io,resources=assessmentprofiles/status,verbs=get;update;patch

func (r *AssessmentProfileReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the AssessmentProfile
	profile := &assessmentv1alpha1.AssessmentProfile{}
	if err := r.Get(ctx, req.NamespacedName, profile); err != nil {
		if client.IgnoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Validate the profile and update status
	ready, message, validatorCount := r.validateProfile(profile)

	// Update status if changed
	if profile.Status.Ready != ready || profile.Status.Message != message || profile.Status.ResolvedValidatorCount != validatorCount {
		profile.Status.Ready = ready
		profile.Status.Message = message
		profile.Status.ResolvedValidatorCount = validatorCount

		if err := r.Status().Update(ctx, profile); err != nil {
			logger.Error(err, "Failed to update AssessmentProfile status")
			return ctrl.Result{}, err
		}
		logger.Info("Updated AssessmentProfile status", "name", profile.Name, "ready", ready, "validators", validatorCount)
	}

	return ctrl.Result{}, nil
}

// validateProfile checks that the AssessmentProfile is valid and returns
// the ready state, a message, and the resolved validator count.
func (r *AssessmentProfileReconciler) validateProfile(profile *assessmentv1alpha1.AssessmentProfile) (bool, string, int) {
	// Validate basedOn
	basedOn := profile.Spec.BasedOn
	if basedOn == "" {
		basedOn = "production"
	}
	if basedOn != string(profiles.ProfileProduction) && basedOn != string(profiles.ProfileDevelopment) {
		return false, fmt.Sprintf("invalid basedOn value %q: must be \"production\" or \"development\"", basedOn), 0
	}

	registeredNames := r.Registry.Names()
	registeredSet := make(map[string]bool, len(registeredNames))
	for _, name := range registeredNames {
		registeredSet[name] = true
	}

	// Validate enabledValidators
	for _, name := range profile.Spec.EnabledValidators {
		if !registeredSet[name] {
			return false, fmt.Sprintf("unknown validator %q in enabledValidators", name), 0
		}
	}

	// Validate disabledValidators
	for _, name := range profile.Spec.DisabledValidators {
		if !registeredSet[name] {
			return false, fmt.Sprintf("unknown validator %q in disabledValidators", name), 0
		}
	}

	// Calculate resolved validator count
	validatorCount := len(registeredNames)
	if len(profile.Spec.EnabledValidators) > 0 {
		validatorCount = len(profile.Spec.EnabledValidators)
	} else if len(profile.Spec.DisabledValidators) > 0 {
		disabledSet := make(map[string]bool, len(profile.Spec.DisabledValidators))
		for _, name := range profile.Spec.DisabledValidators {
			disabledSet[name] = true
		}
		count := 0
		for _, name := range registeredNames {
			if !disabledSet[name] {
				count++
			}
		}
		validatorCount = count
	}

	return true, "Profile is valid", validatorCount
}

// SetupWithManager sets up the controller with the Manager.
func (r *AssessmentProfileReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&assessmentv1alpha1.AssessmentProfile{}).
		Complete(r)
}
