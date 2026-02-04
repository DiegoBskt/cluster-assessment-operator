package controllers

import (
	"fmt"
	"testing"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
)

func BenchmarkFilterBySeverity(b *testing.B) {
	r := &ClusterAssessmentReconciler{}

	// Create a large list of findings
	count := 1000
	findings := make([]assessmentv1alpha1.Finding, count)
	for i := 0; i < count; i++ {
		status := assessmentv1alpha1.FindingStatusInfo
		switch i % 4 {
		case 0:
			status = assessmentv1alpha1.FindingStatusPass
		case 1:
			status = assessmentv1alpha1.FindingStatusWarn
		case 2:
			status = assessmentv1alpha1.FindingStatusFail
		case 3:
			status = assessmentv1alpha1.FindingStatusInfo
		}
		findings[i] = assessmentv1alpha1.Finding{
			ID:     fmt.Sprintf("finding-%d", i),
			Status: status,
			Title:  fmt.Sprintf("Finding %d", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Filter for WARN, which should return WARN and FAIL (half of the findings)
		_ = r.filterBySeverity(findings, "WARN")
	}
}
