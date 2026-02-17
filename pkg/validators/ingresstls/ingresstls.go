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

package ingresstls

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/profiles"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/validator"
)

const (
	validatorName        = "ingresstls"
	validatorDescription = "Validates TLS configuration on Routes and Ingresses"
	validatorCategory    = "Networking"
)

func init() {
	_ = validator.Register(&IngressTLSValidator{})
}

// IngressTLSValidator checks TLS configuration on Routes and Ingresses.
type IngressTLSValidator struct{}

func (v *IngressTLSValidator) Name() string        { return validatorName }
func (v *IngressTLSValidator) Description() string { return validatorDescription }
func (v *IngressTLSValidator) Category() string    { return validatorCategory }

// Validate performs Ingress/Route TLS checks.
func (v *IngressTLSValidator) Validate(ctx context.Context, c client.Client, profile profiles.Profile) ([]assessmentv1alpha1.Finding, error) {
	var findings []assessmentv1alpha1.Finding

	// Check OpenShift Routes
	findings = append(findings, v.checkRoutes(ctx, c, profile)...)

	// Check Kubernetes Ingresses
	findings = append(findings, v.checkIngresses(ctx, c, profile)...)

	return findings, nil
}

// checkRoutes checks OpenShift Routes for TLS configuration.
func (v *IngressTLSValidator) checkRoutes(ctx context.Context, c client.Client, profile profiles.Profile) []assessmentv1alpha1.Finding {
	var findings []assessmentv1alpha1.Finding

	routeList := &unstructured.UnstructuredList{}
	routeList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "route.openshift.io",
		Version: "v1",
		Kind:    "RouteList",
	})

	if err := c.List(ctx, routeList); err != nil {
		// Routes API may not be available (e.g., plain Kubernetes)
		return nil
	}

	var noTLSRoutes []string
	var edgeRoutes int
	var passthroughRoutes int
	var reencryptRoutes int
	totalUserRoutes := 0

	for _, route := range routeList.Items {
		ns := route.GetNamespace()
		name := route.GetName()

		// Skip system namespaces
		if isSystemNamespace(ns) {
			continue
		}
		totalUserRoutes++

		// Check TLS configuration
		tls, found, _ := unstructured.NestedMap(route.Object, "spec", "tls")
		if !found || tls == nil {
			noTLSRoutes = append(noTLSRoutes, fmt.Sprintf("%s/%s", ns, name))
			continue
		}

		termination, _, _ := unstructured.NestedString(route.Object, "spec", "tls", "termination")
		switch termination {
		case "edge":
			edgeRoutes++
		case "passthrough":
			passthroughRoutes++
		case "reencrypt":
			reencryptRoutes++
		}
	}

	if totalUserRoutes == 0 {
		return findings
	}

	// Report routes without TLS
	if len(noTLSRoutes) > 0 {
		status := assessmentv1alpha1.FindingStatusWarn
		if profile.Strictness >= 8 {
			status = assessmentv1alpha1.FindingStatusFail
		}

		sample := noTLSRoutes
		if len(sample) > 10 {
			sample = sample[:10]
		}

		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "ingresstls-routes-no-tls",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         status,
			Title:          "Routes Without TLS",
			Description:    fmt.Sprintf("%d of %d user Routes have no TLS configuration: %s", len(noTLSRoutes), totalUserRoutes, strings.Join(sample, ", ")),
			Impact:         "Traffic to these routes is unencrypted, exposing data in transit.",
			Recommendation: "Enable TLS termination (edge, passthrough, or re-encrypt) on all Routes.",
			References: []string{
				"https://docs.openshift.com/container-platform/latest/networking/routes/secured-routes.html",
			},
			Remediation: &assessmentv1alpha1.RemediationGuidance{
				Safety: assessmentv1alpha1.RemediationRequiresReview,
				Commands: []assessmentv1alpha1.RemediationCommand{
					{Command: "oc get routes -A -o json | jq '.items[] | select(.spec.tls == null) | .metadata.namespace + \"/\" + .metadata.name'", Description: "List all routes without TLS"},
					{Command: "oc patch route <route-name> -n <namespace> --type=merge -p '{\"spec\":{\"tls\":{\"termination\":\"edge\"}}}'", Description: "Enable edge TLS termination", RequiresConfirmation: true},
				},
				DocumentationURL: "https://docs.openshift.com/container-platform/latest/networking/routes/secured-routes.html",
				EstimatedImpact:  "Clients must use HTTPS after enabling TLS",
			},
		})
	} else {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:          "ingresstls-routes-all-tls",
			Validator:   validatorName,
			Category:    validatorCategory,
			Status:      assessmentv1alpha1.FindingStatusPass,
			Title:       "All Routes Have TLS Configured",
			Description: fmt.Sprintf("All %d user Routes have TLS configured (edge: %d, passthrough: %d, re-encrypt: %d).", totalUserRoutes, edgeRoutes, passthroughRoutes, reencryptRoutes),
		})
	}

	return findings
}

// checkIngresses checks Kubernetes Ingresses for TLS configuration.
func (v *IngressTLSValidator) checkIngresses(ctx context.Context, c client.Client, profile profiles.Profile) []assessmentv1alpha1.Finding {
	var findings []assessmentv1alpha1.Finding

	ingressList := &unstructured.UnstructuredList{}
	ingressList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "networking.k8s.io",
		Version: "v1",
		Kind:    "IngressList",
	})

	if err := c.List(ctx, ingressList); err != nil {
		return nil
	}

	var noTLSIngresses []string
	totalUserIngresses := 0

	for _, ingress := range ingressList.Items {
		ns := ingress.GetNamespace()
		name := ingress.GetName()

		if isSystemNamespace(ns) {
			continue
		}
		totalUserIngresses++

		// Check for TLS spec
		tlsList, found, _ := unstructured.NestedSlice(ingress.Object, "spec", "tls")
		if !found || len(tlsList) == 0 {
			noTLSIngresses = append(noTLSIngresses, fmt.Sprintf("%s/%s", ns, name))
		}
	}

	if totalUserIngresses == 0 {
		return findings
	}

	if len(noTLSIngresses) > 0 {
		sample := noTLSIngresses
		if len(sample) > 10 {
			sample = sample[:10]
		}

		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "ingresstls-ingress-no-tls",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         assessmentv1alpha1.FindingStatusWarn,
			Title:          "Ingresses Without TLS",
			Description:    fmt.Sprintf("%d of %d user Ingresses have no TLS configuration: %s", len(noTLSIngresses), totalUserIngresses, strings.Join(sample, ", ")),
			Impact:         "Traffic through these Ingresses is unencrypted.",
			Recommendation: "Configure TLS for all Ingresses with a valid certificate.",
			References: []string{
				"https://kubernetes.io/docs/concepts/services-networking/ingress/#tls",
			},
		})
	} else {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:          "ingresstls-ingress-all-tls",
			Validator:   validatorName,
			Category:    validatorCategory,
			Status:      assessmentv1alpha1.FindingStatusPass,
			Title:       "All Ingresses Have TLS Configured",
			Description: fmt.Sprintf("All %d user Ingresses have TLS configured.", totalUserIngresses),
		})
	}

	return findings
}

func isSystemNamespace(name string) bool {
	return strings.HasPrefix(name, "openshift-") ||
		strings.HasPrefix(name, "kube-") ||
		name == "default" ||
		name == "openshift"
}
