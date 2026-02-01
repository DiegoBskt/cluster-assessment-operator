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
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
)

func TestExportToGit_SecretNamespace(t *testing.T) {
	// Setup scheme
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = assessmentv1alpha1.AddToScheme(scheme)

	// Define resources
	namespace := "tenant-a"
	secretName := "git-creds"

	// Secret in the tenant namespace
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			"username": []byte("user"),
			"password": []byte("pass"),
		},
	}

	// Assessment in the tenant namespace
	assessment := &assessmentv1alpha1.ClusterAssessment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-assessment",
			Namespace: namespace,
		},
		Spec: assessmentv1alpha1.ClusterAssessmentSpec{
			Profile: "default",
			ReportStorage: assessmentv1alpha1.ReportStorageSpec{
				Git: &assessmentv1alpha1.GitStorageSpec{
					Enabled:   true,
					URL:       "https://github.com/example/repo.git",
					SecretRef: secretName,
					Branch:    "main",
				},
			},
		},
	}

	// Create fake client with the secret in the tenant namespace
	// We DO NOT put the secret in the operator namespace
	cl := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(secret, assessment).Build()

	r := &ClusterAssessmentReconciler{
		Client: cl,
		Scheme: scheme,
	}

	// Run exportToGit
	// The function should try to find the secret.
	// Current behavior (BUG): Looks in operator namespace, fails to find secret -> returns "failed to get git secret"
	// Desired behavior (FIX): Looks in tenant namespace, finds secret, fails to clone -> returns "failed to clone repository"
	err := r.exportToGit(context.Background(), assessment)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	errStr := err.Error()
	t.Logf("Got error: %v", errStr)

	// Check if it's the "secret not found" error
	if strings.Contains(errStr, "failed to get git secret") {
		t.Errorf("FAIL: Secret not found. The operator likely looked in the wrong namespace. Error: %v", errStr)
	} else if strings.Contains(errStr, "failed to clone repository") {
		t.Log("Correct behavior: Secret found, proceeded to clone (clone failed as expected due to fake URL)")
	} else {
		t.Errorf("Unexpected error: %v", errStr)
	}
}
