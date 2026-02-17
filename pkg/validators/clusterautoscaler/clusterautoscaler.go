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

package clusterautoscaler

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/profiles"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/validator"
)

const (
	validatorName        = "clusterautoscaler"
	validatorDescription = "Validates Cluster Autoscaler and MachineAutoscaler configuration"
	validatorCategory    = "Platform"
)

func init() {
	_ = validator.Register(&ClusterAutoscalerValidator{})
}

// ClusterAutoscalerValidator checks autoscaler configuration.
type ClusterAutoscalerValidator struct{}

func (v *ClusterAutoscalerValidator) Name() string        { return validatorName }
func (v *ClusterAutoscalerValidator) Description() string { return validatorDescription }
func (v *ClusterAutoscalerValidator) Category() string    { return validatorCategory }

// Validate checks for ClusterAutoscaler and MachineAutoscaler presence.
func (v *ClusterAutoscalerValidator) Validate(ctx context.Context, c client.Client, profile profiles.Profile) ([]assessmentv1alpha1.Finding, error) {
	var findings []assessmentv1alpha1.Finding

	// Check ClusterAutoscaler
	findings = append(findings, v.checkClusterAutoscaler(ctx, c, profile)...)

	// Check MachineAutoscalers
	findings = append(findings, v.checkMachineAutoscalers(ctx, c)...)

	// Check MachineSets for scaling info
	findings = append(findings, v.checkMachineSets(ctx, c)...)

	return findings, nil
}

func (v *ClusterAutoscalerValidator) checkClusterAutoscaler(ctx context.Context, c client.Client, profile profiles.Profile) []assessmentv1alpha1.Finding {
	var findings []assessmentv1alpha1.Finding

	caList := &unstructured.UnstructuredList{}
	caList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "autoscaling.openshift.io",
		Version: "v1",
		Kind:    "ClusterAutoscalerList",
	})

	if err := c.List(ctx, caList); err != nil {
		// API not available — autoscaling not installed, skip gracefully
		return nil
	}

	if len(caList.Items) == 0 {
		status := assessmentv1alpha1.FindingStatusInfo
		if profile.Strictness >= 7 {
			status = assessmentv1alpha1.FindingStatusWarn
		}

		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "autoscaler-no-cluster-autoscaler",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         status,
			Title:          "No ClusterAutoscaler Configured",
			Description:    "No ClusterAutoscaler CR was found. The cluster cannot automatically scale nodes.",
			Impact:         "Without autoscaling, workload spikes may cause resource pressure or scheduling failures.",
			Recommendation: "Consider configuring a ClusterAutoscaler if workloads have variable resource demands.",
			References: []string{
				"https://docs.openshift.com/container-platform/latest/machine_management/applying-autoscaling.html",
			},
			Remediation: &assessmentv1alpha1.RemediationGuidance{
				Safety: assessmentv1alpha1.RemediationRequiresReview,
				Commands: []assessmentv1alpha1.RemediationCommand{
					{Command: "oc get clusterautoscaler", Description: "Check for existing ClusterAutoscaler"},
				},
				DocumentationURL: "https://docs.openshift.com/container-platform/latest/machine_management/applying-autoscaling.html",
				EstimatedImpact:  "Enables automatic node scaling based on workload demand",
			},
		})
	} else {
		ca := caList.Items[0]
		name := ca.GetName()

		// Check resource limits
		maxNodes, _, _ := unstructured.NestedInt64(ca.Object, "spec", "resourceLimits", "maxNodesTotal")
		maxMemory, _, _ := unstructured.NestedInt64(ca.Object, "spec", "resourceLimits", "memory", "maxMemoryTotal")
		maxCores, _, _ := unstructured.NestedInt64(ca.Object, "spec", "resourceLimits", "cores", "maxCoresTotal")

		desc := fmt.Sprintf("ClusterAutoscaler '%s' is configured", name)
		if maxNodes > 0 {
			desc += fmt.Sprintf(" (maxNodes: %d", maxNodes)
			if maxMemory > 0 {
				desc += fmt.Sprintf(", maxMemory: %dGi", maxMemory)
			}
			if maxCores > 0 {
				desc += fmt.Sprintf(", maxCores: %d", maxCores)
			}
			desc += ")"
		}

		findings = append(findings, assessmentv1alpha1.Finding{
			ID:          "autoscaler-cluster-autoscaler-found",
			Validator:   validatorName,
			Category:    validatorCategory,
			Status:      assessmentv1alpha1.FindingStatusPass,
			Title:       "ClusterAutoscaler Configured",
			Description: desc,
		})
	}

	return findings
}

func (v *ClusterAutoscalerValidator) checkMachineAutoscalers(ctx context.Context, c client.Client) []assessmentv1alpha1.Finding {
	var findings []assessmentv1alpha1.Finding

	maList := &unstructured.UnstructuredList{}
	maList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "autoscaling.openshift.io",
		Version: "v1beta1",
		Kind:    "MachineAutoscalerList",
	})

	if err := c.List(ctx, maList); err != nil {
		return nil
	}

	if len(maList.Items) == 0 {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "autoscaler-no-machine-autoscaler",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         assessmentv1alpha1.FindingStatusInfo,
			Title:          "No MachineAutoscalers Found",
			Description:    "No MachineAutoscaler CRs found. Individual MachineSets are not set up for autoscaling.",
			Recommendation: "Create MachineAutoscaler CRs to enable scaling for specific MachineSets.",
		})
	} else {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:          "autoscaler-machine-autoscalers-found",
			Validator:   validatorName,
			Category:    validatorCategory,
			Status:      assessmentv1alpha1.FindingStatusPass,
			Title:       "MachineAutoscalers Configured",
			Description: fmt.Sprintf("%d MachineAutoscaler(s) are configured for individual MachineSets.", len(maList.Items)),
		})
	}

	return findings
}

func (v *ClusterAutoscalerValidator) checkMachineSets(ctx context.Context, c client.Client) []assessmentv1alpha1.Finding {
	var findings []assessmentv1alpha1.Finding

	msList := &unstructured.UnstructuredList{}
	msList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "machine.openshift.io",
		Version: "v1beta1",
		Kind:    "MachineSetList",
	})

	if err := c.List(ctx, msList); err != nil {
		return nil
	}

	zeroReplicas := 0
	for _, ms := range msList.Items {
		replicas, found, _ := unstructured.NestedInt64(ms.Object, "spec", "replicas")
		if found && replicas == 0 {
			zeroReplicas++
		}
	}

	if zeroReplicas > 0 {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "autoscaler-machinesets-zero-replicas",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         assessmentv1alpha1.FindingStatusInfo,
			Title:          "MachineSets With Zero Replicas",
			Description:    fmt.Sprintf("%d MachineSet(s) have 0 replicas. These may be unused or waiting for autoscaling.", zeroReplicas),
			Recommendation: "Review if zero-replica MachineSets are intentional or should be cleaned up.",
		})
	}

	return findings
}
