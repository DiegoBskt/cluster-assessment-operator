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

package rbacaudit

import (
	"context"
	"fmt"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/profiles"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/validator"
)

const (
	validatorName        = "rbacaudit"
	validatorDescription = "Audits namespace-scoped RBAC for overly permissive or risky patterns"
	validatorCategory    = "Security"
)

// Dangerous verbs that grant ability to escalate privileges
var dangerousVerbs = map[string]string{
	"escalate":    "Allows escalating roles beyond the caller's own permissions",
	"bind":        "Allows binding roles to subjects, potentially granting elevated access",
	"impersonate": "Allows impersonating other users, groups, or service accounts",
}

// Sensitive resources that should be tightly controlled
var sensitiveResources = map[string]bool{
	"secrets":                             true,
	"serviceaccounts/token":               true,
	"pods/exec":                           true,
	"pods/attach":                         true,
	"certificatesigningrequests/approval": true,
}

func init() {
	_ = validator.Register(&RBACauditValidator{})
}

// RBACauditValidator audits namespace-scoped RBAC patterns.
type RBACauditValidator struct{}

func (v *RBACauditValidator) Name() string        { return validatorName }
func (v *RBACauditValidator) Description() string { return validatorDescription }
func (v *RBACauditValidator) Category() string    { return validatorCategory }

// Validate performs namespace-scoped RBAC auditing.
func (v *RBACauditValidator) Validate(ctx context.Context, c client.Client, profile profiles.Profile) ([]assessmentv1alpha1.Finding, error) {
	var findings []assessmentv1alpha1.Finding

	// Check 1: Namespace-scoped RoleBindings to cluster-admin
	findings = append(findings, v.checkNamespaceClusterAdminBindings(ctx, c)...)

	// Check 2: Roles with dangerous escalation verbs
	findings = append(findings, v.checkDangerousVerbs(ctx, c)...)

	// Check 3: Roles with wildcard access to sensitive resources
	findings = append(findings, v.checkSensitiveResourceAccess(ctx, c)...)

	// Check 4: RoleBindings with overly broad bindings (all ServiceAccounts)
	findings = append(findings, v.checkBroadBindings(ctx, c)...)

	return findings, nil
}

// checkNamespaceClusterAdminBindings checks for RoleBindings that reference cluster-admin.
func (v *RBACauditValidator) checkNamespaceClusterAdminBindings(ctx context.Context, c client.Client) []assessmentv1alpha1.Finding {
	var findings []assessmentv1alpha1.Finding

	rbs := &rbacv1.RoleBindingList{}
	if err := c.List(ctx, rbs); err != nil {
		return nil
	}

	var clusterAdminRBs []string
	for _, rb := range rbs.Items {
		if isSystemNamespace(rb.Namespace) {
			continue
		}
		if rb.RoleRef.Kind == "ClusterRole" && rb.RoleRef.Name == "cluster-admin" {
			clusterAdminRBs = append(clusterAdminRBs, fmt.Sprintf("%s/%s", rb.Namespace, rb.Name))
		}
	}

	if len(clusterAdminRBs) > 0 {
		sample := clusterAdminRBs
		if len(sample) > 10 {
			sample = sample[:10]
		}

		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "rbacaudit-ns-cluster-admin",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         assessmentv1alpha1.FindingStatusWarn,
			Title:          "Namespace RoleBindings to cluster-admin",
			Description:    fmt.Sprintf("%d RoleBinding(s) in user namespaces grant cluster-admin access: %s", len(clusterAdminRBs), strings.Join(sample, ", ")),
			Impact:         "Namespace-scoped cluster-admin bindings grant full cluster privileges within that namespace, defeating namespace isolation.",
			Recommendation: "Replace cluster-admin references with more specific Roles scoped to the namespace needs.",
			Remediation: &assessmentv1alpha1.RemediationGuidance{
				Safety: assessmentv1alpha1.RemediationRequiresReview,
				Commands: []assessmentv1alpha1.RemediationCommand{
					{Command: "oc get rolebindings -A -o json | jq '.items[] | select(.roleRef.name==\"cluster-admin\") | .metadata.namespace + \"/\" + .metadata.name'", Description: "List all namespace RoleBindings with cluster-admin"},
				},
				EstimatedImpact: "Users may lose access if RoleBindings are changed",
			},
		})
	} else {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:          "rbacaudit-no-ns-cluster-admin",
			Validator:   validatorName,
			Category:    validatorCategory,
			Status:      assessmentv1alpha1.FindingStatusPass,
			Title:       "No Namespace RoleBindings to cluster-admin",
			Description: "No namespace-scoped RoleBindings reference cluster-admin.",
		})
	}

	return findings
}

// checkDangerousVerbs checks for Roles/ClusterRoles with escalation verbs.
func (v *RBACauditValidator) checkDangerousVerbs(ctx context.Context, c client.Client) []assessmentv1alpha1.Finding {
	var findings []assessmentv1alpha1.Finding

	// Check ClusterRoles
	clusterRoles := &rbacv1.ClusterRoleList{}
	if err := c.List(ctx, clusterRoles); err != nil {
		return nil
	}

	var escalationRoles []string
	for _, cr := range clusterRoles.Items {
		if strings.HasPrefix(cr.Name, "system:") || strings.HasPrefix(cr.Name, "openshift") {
			continue
		}
		for _, rule := range cr.Rules {
			for _, verb := range rule.Verbs {
				if _, isDangerous := dangerousVerbs[verb]; isDangerous {
					escalationRoles = append(escalationRoles, fmt.Sprintf("ClusterRole/%s (verb: %s)", cr.Name, verb))
					break
				}
			}
		}
	}

	// Check namespace Roles
	roles := &rbacv1.RoleList{}
	if err := c.List(ctx, roles); err != nil {
		return nil
	}

	for _, role := range roles.Items {
		if isSystemNamespace(role.Namespace) {
			continue
		}
		for _, rule := range role.Rules {
			for _, verb := range rule.Verbs {
				if _, isDangerous := dangerousVerbs[verb]; isDangerous {
					escalationRoles = append(escalationRoles, fmt.Sprintf("Role/%s/%s (verb: %s)", role.Namespace, role.Name, verb))
					break
				}
			}
		}
	}

	// Deduplicate
	escalationRoles = unique(escalationRoles)

	if len(escalationRoles) > 0 {
		sample := escalationRoles
		if len(sample) > 10 {
			sample = sample[:10]
		}

		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "rbacaudit-dangerous-verbs",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         assessmentv1alpha1.FindingStatusWarn,
			Title:          "Roles With Privilege Escalation Verbs",
			Description:    fmt.Sprintf("%d Role(s) have escalation verbs (escalate, bind, impersonate): %s", len(escalationRoles), strings.Join(sample, "; ")),
			Impact:         "These verbs allow users to grant themselves or others additional permissions beyond what they currently have.",
			Recommendation: "Review and restrict escalation verbs to only trusted administrative roles.",
			References: []string{
				"https://kubernetes.io/docs/reference/access-authn-authz/rbac/#privilege-escalation-prevention-and-bootstrapping",
			},
		})
	} else {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:          "rbacaudit-no-dangerous-verbs",
			Validator:   validatorName,
			Category:    validatorCategory,
			Status:      assessmentv1alpha1.FindingStatusPass,
			Title:       "No Custom Roles With Escalation Verbs",
			Description: "No custom Roles or ClusterRoles use escalate, bind, or impersonate verbs.",
		})
	}

	return findings
}

// checkSensitiveResourceAccess checks for Roles with wildcard or broad access to sensitive resources.
func (v *RBACauditValidator) checkSensitiveResourceAccess(ctx context.Context, c client.Client) []assessmentv1alpha1.Finding {
	var findings []assessmentv1alpha1.Finding

	roles := &rbacv1.RoleList{}
	if err := c.List(ctx, roles); err != nil {
		return nil
	}

	var sensitiveAccess []string
	for _, role := range roles.Items {
		if isSystemNamespace(role.Namespace) {
			continue
		}
		for _, rule := range role.Rules {
			for _, resource := range rule.Resources {
				if sensitiveResources[resource] || resource == "*" {
					for _, verb := range rule.Verbs {
						if verb == "*" || verb == "create" || verb == "update" || verb == "patch" || verb == "delete" {
							sensitiveAccess = append(sensitiveAccess, fmt.Sprintf("%s/%s (%s on %s)", role.Namespace, role.Name, verb, resource))
							break
						}
					}
				}
			}
		}
	}

	sensitiveAccess = unique(sensitiveAccess)

	if len(sensitiveAccess) > 0 {
		sample := sensitiveAccess
		if len(sample) > 10 {
			sample = sample[:10]
		}

		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "rbacaudit-sensitive-access",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         assessmentv1alpha1.FindingStatusInfo,
			Title:          "Roles With Write Access to Sensitive Resources",
			Description:    fmt.Sprintf("%d Role(s) grant write access to sensitive resources (secrets, exec, token): %s", len(sensitiveAccess), strings.Join(sample, "; ")),
			Impact:         "Write access to sensitive resources can be used to extract credentials or execute arbitrary commands in pods.",
			Recommendation: "Restrict write access to sensitive resources to only the roles that strictly require it.",
		})
	}

	return findings
}

// checkBroadBindings checks for RoleBindings that bind to all ServiceAccounts in a namespace.
func (v *RBACauditValidator) checkBroadBindings(ctx context.Context, c client.Client) []assessmentv1alpha1.Finding {
	var findings []assessmentv1alpha1.Finding

	rbs := &rbacv1.RoleBindingList{}
	if err := c.List(ctx, rbs); err != nil {
		return nil
	}

	var broadBindings []string
	for _, rb := range rbs.Items {
		if isSystemNamespace(rb.Namespace) {
			continue
		}
		for _, subject := range rb.Subjects {
			if subject.Kind == "Group" && subject.Name == "system:serviceaccounts" {
				broadBindings = append(broadBindings, fmt.Sprintf("%s/%s", rb.Namespace, rb.Name))
			}
		}
	}

	if len(broadBindings) > 0 {
		sample := broadBindings
		if len(sample) > 10 {
			sample = sample[:10]
		}

		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "rbacaudit-broad-bindings",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         assessmentv1alpha1.FindingStatusWarn,
			Title:          "RoleBindings Granting Access to All Service Accounts",
			Description:    fmt.Sprintf("%d RoleBinding(s) bind to the system:serviceaccounts group, granting permissions to every ServiceAccount: %s", len(broadBindings), strings.Join(sample, ", ")),
			Impact:         "Any pod in any namespace can inherit permissions from these bindings via its service account.",
			Recommendation: "Bind to specific ServiceAccounts instead of the broad system:serviceaccounts group.",
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

func unique(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
