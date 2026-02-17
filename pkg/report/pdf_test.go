package report

import (
	"os"
	"testing"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGeneratePDFWithManyFindings(t *testing.T) {
	score := 72
	assessment := &assessmentv1alpha1.ClusterAssessment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-assessment",
		},
		Status: assessmentv1alpha1.ClusterAssessmentStatus{
			ClusterInfo: assessmentv1alpha1.ClusterInfo{
				ClusterID:      "test-cluster-123",
				ClusterVersion: "4.20.1",
				Platform:       "AWS",
				Channel:        "stable-4.20",
				NodeCount:      6,
			},
			Summary: assessmentv1alpha1.AssessmentSummary{
				Score:       &score,
				TotalChecks: 60,
				PassCount:   40,
				WarnCount:   10,
				FailCount:   8,
				InfoCount:   2,
			},
			Findings: []assessmentv1alpha1.Finding{
				{Title: "Security Finding 1", Description: "A security issue found", Category: "Security", Validator: "security", Status: assessmentv1alpha1.FindingStatusFail, Recommendation: "Fix this"},
				{Title: "Security Finding 2", Description: "PSA labels missing", Category: "Security", Validator: "podsecurityadmission", Status: assessmentv1alpha1.FindingStatusWarn, Recommendation: "Add labels"},
				{Title: "Security Finding 3", Description: "RBAC audit issue", Category: "Security", Validator: "rbacaudit", Status: assessmentv1alpha1.FindingStatusFail},
				{Title: "Security Finding 4", Description: "Certificate expiring", Category: "Security", Validator: "certificates", Status: assessmentv1alpha1.FindingStatusWarn},
				{Title: "Security Pass", Description: "Good security config", Category: "Security", Validator: "security", Status: assessmentv1alpha1.FindingStatusPass},
				{Title: "Platform Finding 1", Description: "Version check", Category: "Platform", Validator: "version", Status: assessmentv1alpha1.FindingStatusPass},
				{Title: "Platform Finding 2", Description: "ETCD backup missing", Category: "Platform", Validator: "etcdbackup", Status: assessmentv1alpha1.FindingStatusFail},
				{Title: "Platform Finding 3", Description: "API server ok", Category: "Platform", Validator: "apiserver", Status: assessmentv1alpha1.FindingStatusPass},
				{Title: "Platform Finding 4", Description: "Operator degraded", Category: "Platform", Validator: "operators", Status: assessmentv1alpha1.FindingStatusWarn},
				{Title: "Platform Finding 5", Description: "Autoscaler missing", Category: "Platform", Validator: "clusterautoscaler", Status: assessmentv1alpha1.FindingStatusInfo},
				{Title: "Platform Finding 6", Description: "OADP backup stale", Category: "Platform", Validator: "oadpbackup", Status: assessmentv1alpha1.FindingStatusWarn},
				{Title: "Networking Finding 1", Description: "Missing TLS on routes", Category: "Networking", Validator: "ingresstls", Status: assessmentv1alpha1.FindingStatusFail},
				{Title: "Networking Finding 2", Description: "Network policy ok", Category: "Networking", Validator: "networking", Status: assessmentv1alpha1.FindingStatusPass},
				{Title: "Networking Finding 3", Description: "NPAudit issue", Category: "Networking", Validator: "networkpolicyaudit", Status: assessmentv1alpha1.FindingStatusWarn},
				{Title: "Infrastructure 1", Description: "Node count ok", Category: "Infrastructure", Validator: "nodes", Status: assessmentv1alpha1.FindingStatusPass},
				{Title: "Infrastructure 2", Description: "Cost optimization", Category: "Infrastructure", Validator: "costoptimization", Status: assessmentv1alpha1.FindingStatusInfo},
				{Title: "Observability 1", Description: "Monitoring ok", Category: "Observability", Validator: "monitoring", Status: assessmentv1alpha1.FindingStatusPass},
				{Title: "Observability 2", Description: "Logging issue", Category: "Observability", Validator: "logging", Status: assessmentv1alpha1.FindingStatusWarn},
				{Title: "Storage 1", Description: "Storage class ok", Category: "Storage", Validator: "storage", Status: assessmentv1alpha1.FindingStatusPass},
				{Title: "Governance 1", Description: "ResourceQuotas missing", Category: "Governance", Validator: "resourcequotas", Status: assessmentv1alpha1.FindingStatusWarn},
				{Title: "Compatibility 1", Description: "Deprecated API", Category: "Compatibility", Validator: "deprecation", Status: assessmentv1alpha1.FindingStatusFail},
				{Title: "MachineConfig 1", Description: "MCP check", Category: "Platform", Validator: "machineconfig", Status: assessmentv1alpha1.FindingStatusPass},
				{Title: "ImageRegistry 1", Description: "Registry config", Category: "Platform", Validator: "imageregistry", Status: assessmentv1alpha1.FindingStatusPass},
			},
		},
	}

	data, err := GeneratePDF(assessment)
	if err != nil {
		t.Fatalf("Failed to generate PDF: %v", err)
	}

	if err := os.WriteFile("/tmp/test-assessment.pdf", data, 0644); err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}

	t.Logf("PDF generated: /tmp/test-assessment.pdf (%d bytes)", len(data))
}
