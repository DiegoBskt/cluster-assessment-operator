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

package podsecurityadmission

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/profiles"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/validator"
)

const (
	validatorName        = "podsecurityadmission"
	validatorDescription = "Validates Pod Security Admission labels on namespaces"
	validatorCategory    = "Security"
)

// PSA label keys
const (
	psaEnforce = "pod-security.kubernetes.io/enforce"
	psaWarn    = "pod-security.kubernetes.io/warn"
	psaAudit   = "pod-security.kubernetes.io/audit"
)

func init() {
	_ = validator.Register(&PSAValidator{})
}

// PSAValidator checks Pod Security Admission configuration.
type PSAValidator struct{}

func (v *PSAValidator) Name() string        { return validatorName }
func (v *PSAValidator) Description() string { return validatorDescription }
func (v *PSAValidator) Category() string    { return validatorCategory }

// Validate checks PSA labels on all user namespaces.
func (v *PSAValidator) Validate(ctx context.Context, c client.Client, profile profiles.Profile) ([]assessmentv1alpha1.Finding, error) {
	var findings []assessmentv1alpha1.Finding

	namespaces := &corev1.NamespaceList{}
	if err := c.List(ctx, namespaces); err != nil {
		return []assessmentv1alpha1.Finding{{
			ID:          "psa-list-error",
			Validator:   validatorName,
			Category:    validatorCategory,
			Status:      assessmentv1alpha1.FindingStatusFail,
			Title:       "Unable to List Namespaces",
			Description: fmt.Sprintf("Failed to list namespaces: %v", err),
		}}, nil
	}

	var noPSALabels []string
	var privilegedEnforce []string
	var restrictedEnforce []string
	totalUser := 0

	for _, ns := range namespaces.Items {
		// Skip system namespaces
		if isSystemNamespace(ns.Name) {
			continue
		}
		totalUser++

		enforce := ns.Labels[psaEnforce]
		warn := ns.Labels[psaWarn]
		audit := ns.Labels[psaAudit]

		if enforce == "" && warn == "" && audit == "" {
			noPSALabels = append(noPSALabels, ns.Name)
		}

		if enforce == "privileged" {
			privilegedEnforce = append(privilegedEnforce, ns.Name)
		}
		if enforce == "restricted" {
			restrictedEnforce = append(restrictedEnforce, ns.Name)
		}
	}

	// Report namespaces without PSA labels
	if len(noPSALabels) > 0 {
		status := assessmentv1alpha1.FindingStatusInfo
		if profile.Strictness >= 7 {
			status = assessmentv1alpha1.FindingStatusWarn
		}

		sample := noPSALabels
		if len(sample) > 10 {
			sample = sample[:10]
		}

		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "psa-no-labels",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         status,
			Title:          "Namespaces Without Pod Security Admission Labels",
			Description:    fmt.Sprintf("%d of %d user namespaces have no PSA labels: %s", len(noPSALabels), totalUser, strings.Join(sample, ", ")),
			Impact:         "Without PSA labels, pods in these namespaces run with the default (usually privileged) security level.",
			Recommendation: "Add pod-security.kubernetes.io/enforce labels to configure the security level for each namespace.",
			References: []string{
				"https://kubernetes.io/docs/concepts/security/pod-security-admission/",
				"https://docs.openshift.com/container-platform/latest/authentication/understanding-and-managing-pod-security-admission.html",
			},
			Remediation: &assessmentv1alpha1.RemediationGuidance{
				Safety: assessmentv1alpha1.RemediationRequiresReview,
				Commands: []assessmentv1alpha1.RemediationCommand{
					{Command: "oc label namespace <namespace> pod-security.kubernetes.io/enforce=restricted", Description: "Set restricted enforcement for a namespace", RequiresConfirmation: true},
					{Command: "oc label namespace <namespace> pod-security.kubernetes.io/warn=restricted", Description: "Set restricted warnings for a namespace"},
					{Command: "oc get namespaces -l '!pod-security.kubernetes.io/enforce' --no-headers | awk '{print $1}'", Description: "List namespaces without enforce label"},
				},
				DocumentationURL: "https://kubernetes.io/docs/concepts/security/pod-security-admission/",
				EstimatedImpact:  "Pods that violate the enforcement level will be rejected",
				Prerequisites:    []string{"Review existing workloads to ensure they comply with the target security level"},
			},
		})
	} else {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:          "psa-all-labeled",
			Validator:   validatorName,
			Category:    validatorCategory,
			Status:      assessmentv1alpha1.FindingStatusPass,
			Title:       "All User Namespaces Have PSA Labels",
			Description: fmt.Sprintf("All %d user namespaces have Pod Security Admission labels configured.", totalUser),
		})
	}

	// Report privileged namespaces
	if len(privilegedEnforce) > 0 {
		status := assessmentv1alpha1.FindingStatusInfo
		if profile.Strictness >= 7 && !profile.Thresholds.AllowPrivilegedContainers {
			status = assessmentv1alpha1.FindingStatusWarn
		}

		sample := privilegedEnforce
		if len(sample) > 10 {
			sample = sample[:10]
		}

		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "psa-privileged-enforce",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         status,
			Title:          "Namespaces With Privileged PSA Enforcement",
			Description:    fmt.Sprintf("%d namespace(s) enforce the 'privileged' Pod Security level: %s", len(privilegedEnforce), strings.Join(sample, ", ")),
			Impact:         "Privileged enforcement allows pods with any security configuration, including host access.",
			Recommendation: "Consider using 'baseline' or 'restricted' enforcement where possible.",
		})
	}

	// Report restricted namespaces (positive finding)
	if len(restrictedEnforce) > 0 {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:          "psa-restricted-enforce",
			Validator:   validatorName,
			Category:    validatorCategory,
			Status:      assessmentv1alpha1.FindingStatusPass,
			Title:       "Namespaces With Restricted PSA Enforcement",
			Description: fmt.Sprintf("%d namespace(s) enforce the 'restricted' Pod Security level, providing strong pod security.", len(restrictedEnforce)),
		})
	}

	return findings, nil
}

func isSystemNamespace(name string) bool {
	return strings.HasPrefix(name, "openshift-") ||
		strings.HasPrefix(name, "kube-") ||
		name == "default" ||
		name == "openshift"
}
