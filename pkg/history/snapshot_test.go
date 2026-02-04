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

package history

import (
	"sort"
	"testing"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func intPtr(i int) *int { return &i }

func TestCompactFindings(t *testing.T) {
	findings := []assessmentv1alpha1.Finding{
		{
			ID:             "security-1",
			Validator:      "security",
			Category:       "Security",
			Status:         assessmentv1alpha1.FindingStatusFail,
			Title:          "Excessive cluster-admin bindings",
			Description:    "Long description...",
			Impact:         "Security impact...",
			Recommendation: "Reduce bindings...",
			Resource:       "my-binding",
			Namespace:      "",
			References:     []string{"https://docs.openshift.com"},
		},
		{
			ID:        "nodes-1",
			Validator: "nodes",
			Category:  "Infrastructure",
			Status:    assessmentv1alpha1.FindingStatusPass,
			Title:     "Node count OK",
		},
	}

	compact := compactFindings(findings)

	if len(compact) != 2 {
		t.Fatalf("Expected 2 compact findings, got %d", len(compact))
	}

	// Check first finding
	if compact[0].ID != "security-1" {
		t.Errorf("Expected ID 'security-1', got %q", compact[0].ID)
	}
	if compact[0].Resource != "my-binding" {
		t.Errorf("Expected Resource 'my-binding', got %q", compact[0].Resource)
	}
	if compact[0].Status != assessmentv1alpha1.FindingStatusFail {
		t.Errorf("Expected status FAIL, got %s", compact[0].Status)
	}

	// Check second finding
	if compact[1].ID != "nodes-1" {
		t.Errorf("Expected ID 'nodes-1', got %q", compact[1].ID)
	}
}

func TestComputeDelta_NilPrevious(t *testing.T) {
	current := []assessmentv1alpha1.FindingSnapshot{
		{ID: "check-1", Status: assessmentv1alpha1.FindingStatusPass},
	}

	delta := ComputeDelta(current, intPtr(80), nil)

	if delta != nil {
		t.Error("Expected nil delta for nil previous")
	}
}

func TestComputeDelta_NewFindings(t *testing.T) {
	current := []assessmentv1alpha1.FindingSnapshot{
		{ID: "check-1", Status: assessmentv1alpha1.FindingStatusPass},
		{ID: "check-2", Status: assessmentv1alpha1.FindingStatusFail},
		{ID: "check-3", Status: assessmentv1alpha1.FindingStatusWarn},
	}

	previous := &assessmentv1alpha1.AssessmentSnapshot{
		Status: assessmentv1alpha1.AssessmentSnapshotStatus{
			Findings: []assessmentv1alpha1.FindingSnapshot{
				{ID: "check-1", Status: assessmentv1alpha1.FindingStatusPass},
			},
			Summary: assessmentv1alpha1.AssessmentSummary{Score: intPtr(90)},
		},
	}

	delta := ComputeDelta(current, intPtr(70), previous)

	if delta == nil {
		t.Fatal("Expected non-nil delta")
		return
	}

	sort.Strings(delta.NewFindings)
	if len(delta.NewFindings) != 2 {
		t.Errorf("Expected 2 new findings, got %d: %v", len(delta.NewFindings), delta.NewFindings)
	}
	if delta.NewFindings[0] != "check-2" || delta.NewFindings[1] != "check-3" {
		t.Errorf("Expected [check-2, check-3], got %v", delta.NewFindings)
	}
}

func TestComputeDelta_ResolvedFindings(t *testing.T) {
	current := []assessmentv1alpha1.FindingSnapshot{
		{ID: "check-1", Status: assessmentv1alpha1.FindingStatusPass},
	}

	previous := &assessmentv1alpha1.AssessmentSnapshot{
		Status: assessmentv1alpha1.AssessmentSnapshotStatus{
			Findings: []assessmentv1alpha1.FindingSnapshot{
				{ID: "check-1", Status: assessmentv1alpha1.FindingStatusPass},
				{ID: "check-2", Status: assessmentv1alpha1.FindingStatusFail},
				{ID: "check-3", Status: assessmentv1alpha1.FindingStatusWarn},
			},
			Summary: assessmentv1alpha1.AssessmentSummary{Score: intPtr(60)},
		},
	}

	delta := ComputeDelta(current, intPtr(90), previous)

	if len(delta.ResolvedFindings) != 2 {
		t.Errorf("Expected 2 resolved findings, got %d: %v", len(delta.ResolvedFindings), delta.ResolvedFindings)
	}
}

func TestComputeDelta_Regressions(t *testing.T) {
	current := []assessmentv1alpha1.FindingSnapshot{
		{ID: "check-1", Status: assessmentv1alpha1.FindingStatusFail}, // was WARN
		{ID: "check-2", Status: assessmentv1alpha1.FindingStatusWarn}, // was PASS
	}

	previous := &assessmentv1alpha1.AssessmentSnapshot{
		Status: assessmentv1alpha1.AssessmentSnapshotStatus{
			Findings: []assessmentv1alpha1.FindingSnapshot{
				{ID: "check-1", Status: assessmentv1alpha1.FindingStatusWarn},
				{ID: "check-2", Status: assessmentv1alpha1.FindingStatusPass},
			},
			Summary: assessmentv1alpha1.AssessmentSummary{Score: intPtr(80)},
		},
	}

	delta := ComputeDelta(current, intPtr(50), previous)

	if len(delta.RegressionFindings) != 2 {
		t.Errorf("Expected 2 regressions, got %d: %v", len(delta.RegressionFindings), delta.RegressionFindings)
	}
	if len(delta.ImprovedFindings) != 0 {
		t.Errorf("Expected 0 improvements, got %d", len(delta.ImprovedFindings))
	}
}

func TestComputeDelta_Improvements(t *testing.T) {
	current := []assessmentv1alpha1.FindingSnapshot{
		{ID: "check-1", Status: assessmentv1alpha1.FindingStatusPass}, // was FAIL
		{ID: "check-2", Status: assessmentv1alpha1.FindingStatusWarn}, // was FAIL
	}

	previous := &assessmentv1alpha1.AssessmentSnapshot{
		Status: assessmentv1alpha1.AssessmentSnapshotStatus{
			Findings: []assessmentv1alpha1.FindingSnapshot{
				{ID: "check-1", Status: assessmentv1alpha1.FindingStatusFail},
				{ID: "check-2", Status: assessmentv1alpha1.FindingStatusFail},
			},
			Summary: assessmentv1alpha1.AssessmentSummary{Score: intPtr(30)},
		},
	}

	delta := ComputeDelta(current, intPtr(70), previous)

	if len(delta.ImprovedFindings) != 2 {
		t.Errorf("Expected 2 improvements, got %d: %v", len(delta.ImprovedFindings), delta.ImprovedFindings)
	}
	if len(delta.RegressionFindings) != 0 {
		t.Errorf("Expected 0 regressions, got %d", len(delta.RegressionFindings))
	}
}

func TestComputeDelta_ScoreDelta(t *testing.T) {
	current := []assessmentv1alpha1.FindingSnapshot{}
	previous := &assessmentv1alpha1.AssessmentSnapshot{
		Status: assessmentv1alpha1.AssessmentSnapshotStatus{
			Summary: assessmentv1alpha1.AssessmentSummary{Score: intPtr(72)},
		},
	}

	delta := ComputeDelta(current, intPtr(85), previous)

	if delta.ScoreDelta == nil {
		t.Fatal("Expected ScoreDelta to be set")
	}
	if *delta.ScoreDelta != 13 {
		t.Errorf("Expected ScoreDelta=13, got %d", *delta.ScoreDelta)
	}
}

func TestComputeDelta_ScoreDeltaNegative(t *testing.T) {
	current := []assessmentv1alpha1.FindingSnapshot{}
	previous := &assessmentv1alpha1.AssessmentSnapshot{
		Status: assessmentv1alpha1.AssessmentSnapshotStatus{
			Summary: assessmentv1alpha1.AssessmentSummary{Score: intPtr(85)},
		},
	}

	delta := ComputeDelta(current, intPtr(72), previous)

	if delta.ScoreDelta == nil {
		t.Fatal("Expected ScoreDelta to be set")
	}
	if *delta.ScoreDelta != -13 {
		t.Errorf("Expected ScoreDelta=-13, got %d", *delta.ScoreDelta)
	}
}

func TestComputeDelta_ScoreDeltaNilScores(t *testing.T) {
	current := []assessmentv1alpha1.FindingSnapshot{}
	previous := &assessmentv1alpha1.AssessmentSnapshot{
		Status: assessmentv1alpha1.AssessmentSnapshotStatus{
			Summary: assessmentv1alpha1.AssessmentSummary{Score: nil},
		},
	}

	delta := ComputeDelta(current, nil, previous)

	if delta.ScoreDelta != nil {
		t.Error("Expected ScoreDelta to be nil when scores are nil")
	}
}

func TestComputeDelta_NoChanges(t *testing.T) {
	findings := []assessmentv1alpha1.FindingSnapshot{
		{ID: "check-1", Status: assessmentv1alpha1.FindingStatusPass},
		{ID: "check-2", Status: assessmentv1alpha1.FindingStatusWarn},
	}

	previous := &assessmentv1alpha1.AssessmentSnapshot{
		Status: assessmentv1alpha1.AssessmentSnapshotStatus{
			Findings: []assessmentv1alpha1.FindingSnapshot{
				{ID: "check-1", Status: assessmentv1alpha1.FindingStatusPass},
				{ID: "check-2", Status: assessmentv1alpha1.FindingStatusWarn},
			},
			Summary: assessmentv1alpha1.AssessmentSummary{Score: intPtr(80)},
		},
	}

	delta := ComputeDelta(findings, intPtr(80), previous)

	if len(delta.NewFindings) != 0 {
		t.Errorf("Expected 0 new findings, got %d", len(delta.NewFindings))
	}
	if len(delta.ResolvedFindings) != 0 {
		t.Errorf("Expected 0 resolved findings, got %d", len(delta.ResolvedFindings))
	}
	if len(delta.RegressionFindings) != 0 {
		t.Errorf("Expected 0 regressions, got %d", len(delta.RegressionFindings))
	}
	if len(delta.ImprovedFindings) != 0 {
		t.Errorf("Expected 0 improvements, got %d", len(delta.ImprovedFindings))
	}
	if *delta.ScoreDelta != 0 {
		t.Errorf("Expected ScoreDelta=0, got %d", *delta.ScoreDelta)
	}
}

func TestComputeDelta_MixedChanges(t *testing.T) {
	current := []assessmentv1alpha1.FindingSnapshot{
		{ID: "check-1", Status: assessmentv1alpha1.FindingStatusPass},  // improved (was WARN)
		{ID: "check-2", Status: assessmentv1alpha1.FindingStatusFail},  // regressed (was WARN)
		{ID: "check-4", Status: assessmentv1alpha1.FindingStatusInfo},  // new
	}

	previous := &assessmentv1alpha1.AssessmentSnapshot{
		Status: assessmentv1alpha1.AssessmentSnapshotStatus{
			Findings: []assessmentv1alpha1.FindingSnapshot{
				{ID: "check-1", Status: assessmentv1alpha1.FindingStatusWarn},
				{ID: "check-2", Status: assessmentv1alpha1.FindingStatusWarn},
				{ID: "check-3", Status: assessmentv1alpha1.FindingStatusFail}, // resolved
			},
			Summary: assessmentv1alpha1.AssessmentSummary{Score: intPtr(60)},
		},
	}

	delta := ComputeDelta(current, intPtr(65), previous)

	if len(delta.NewFindings) != 1 || delta.NewFindings[0] != "check-4" {
		t.Errorf("Expected 1 new finding (check-4), got %v", delta.NewFindings)
	}
	if len(delta.ResolvedFindings) != 1 || delta.ResolvedFindings[0] != "check-3" {
		t.Errorf("Expected 1 resolved finding (check-3), got %v", delta.ResolvedFindings)
	}
	if len(delta.RegressionFindings) != 1 || delta.RegressionFindings[0] != "check-2" {
		t.Errorf("Expected 1 regression (check-2), got %v", delta.RegressionFindings)
	}
	if len(delta.ImprovedFindings) != 1 || delta.ImprovedFindings[0] != "check-1" {
		t.Errorf("Expected 1 improvement (check-1), got %v", delta.ImprovedFindings)
	}
	if *delta.ScoreDelta != 5 {
		t.Errorf("Expected ScoreDelta=5, got %d", *delta.ScoreDelta)
	}
}

func TestSeverityLevel(t *testing.T) {
	tests := []struct {
		status assessmentv1alpha1.FindingStatus
		level  int
	}{
		{assessmentv1alpha1.FindingStatusInfo, 0},
		{assessmentv1alpha1.FindingStatusPass, 1},
		{assessmentv1alpha1.FindingStatusWarn, 2},
		{assessmentv1alpha1.FindingStatusFail, 3},
	}

	for _, tt := range tests {
		got := severityLevel(tt.status)
		if got != tt.level {
			t.Errorf("severityLevel(%s) = %d, want %d", tt.status, got, tt.level)
		}
	}
}

// Ensure metav1 import is used
var _ = metav1.Now
