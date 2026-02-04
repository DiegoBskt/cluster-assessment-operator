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
	"context"
	"fmt"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Resolver resolves a profile from a ClusterAssessment spec.
// It handles both built-in profile names and custom AssessmentProfile CR references.
type Resolver struct {
	client client.Client
}

// NewResolver creates a new profile Resolver.
func NewResolver(c client.Client) *Resolver {
	return &Resolver{client: c}
}

// Resolve returns the effective Profile for a given ClusterAssessment.
// Resolution order:
//  1. If profile name matches a built-in ("production", "development"), return it directly.
//  2. Otherwise, look up an AssessmentProfile CR with that name and merge with its base.
func (r *Resolver) Resolve(ctx context.Context, profileName string) (Profile, error) {
	if profileName == "" {
		profileName = string(ProfileProduction)
	}

	// Check built-in profiles first
	if profileName == string(ProfileProduction) || profileName == string(ProfileDevelopment) {
		return GetProfile(profileName), nil
	}

	// Look up custom AssessmentProfile CR
	customProfile := &assessmentv1alpha1.AssessmentProfile{}
	if err := r.client.Get(ctx, client.ObjectKey{Name: profileName}, customProfile); err != nil {
		return Profile{}, fmt.Errorf("profile %q not found: %w", profileName, err)
	}

	return mergeProfile(customProfile), nil
}

// mergeProfile creates a Profile by starting with the base profile and applying
// overrides from the custom AssessmentProfile. Nil pointer fields in ThresholdOverrides
// are left at base profile defaults.
func mergeProfile(custom *assessmentv1alpha1.AssessmentProfile) Profile {
	baseName := custom.Spec.BasedOn
	if baseName == "" {
		baseName = string(ProfileProduction)
	}
	base := GetProfile(baseName)

	// Override profile identity
	base.Name = ProfileName(custom.Name)
	if custom.Spec.Description != "" {
		base.Description = custom.Spec.Description
	}

	// Merge threshold overrides (nil = inherit from base)
	if t := custom.Spec.Thresholds; t != nil {
		if t.MinControlPlaneNodes != nil {
			base.Thresholds.MinControlPlaneNodes = *t.MinControlPlaneNodes
		}
		if t.MinWorkerNodes != nil {
			base.Thresholds.MinWorkerNodes = *t.MinWorkerNodes
		}
		if t.MaxPodsPerNode != nil {
			base.Thresholds.MaxPodsPerNode = *t.MaxPodsPerNode
		}
		if t.MaxClusterAdminBindings != nil {
			base.Thresholds.MaxClusterAdminBindings = *t.MaxClusterAdminBindings
		}
		if t.RequireNetworkPolicy != nil {
			base.Thresholds.RequireNetworkPolicy = *t.RequireNetworkPolicy
		}
		if t.RequireResourceQuotas != nil {
			base.Thresholds.RequireResourceQuotas = *t.RequireResourceQuotas
		}
		if t.RequireLimitRanges != nil {
			base.Thresholds.RequireLimitRanges = *t.RequireLimitRanges
		}
		if t.MaxDaysWithoutUpdate != nil {
			base.Thresholds.MaxDaysWithoutUpdate = *t.MaxDaysWithoutUpdate
		}
		if t.AllowPrivilegedContainers != nil {
			base.Thresholds.AllowPrivilegedContainers = *t.AllowPrivilegedContainers
		}
		if t.RequireDefaultStorageClass != nil {
			base.Thresholds.RequireDefaultStorageClass = *t.RequireDefaultStorageClass
		}
	}

	// Merge validator lists
	if len(custom.Spec.EnabledValidators) > 0 {
		base.EnabledValidators = custom.Spec.EnabledValidators
	}
	if len(custom.Spec.DisabledChecks) > 0 {
		base.DisabledChecks = append(base.DisabledChecks, custom.Spec.DisabledChecks...)
	}

	return base
}
