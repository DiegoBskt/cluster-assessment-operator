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
	"context"
	"fmt"
	"sort"
	"time"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// LabelAssessmentName is the label key used to link snapshots to assessments.
	LabelAssessmentName = "assessment.openshift.io/name"
)

// SnapshotManager handles creating, querying, and pruning assessment snapshots.
type SnapshotManager struct {
	client client.Client
}

// NewSnapshotManager creates a new SnapshotManager.
func NewSnapshotManager(c client.Client) *SnapshotManager {
	return &SnapshotManager{client: c}
}

// CreateSnapshot creates a new AssessmentSnapshot from a completed assessment.
// It computes the delta from the previous snapshot and prunes old snapshots.
// Returns the created snapshot's delta summary and snapshot count.
func (m *SnapshotManager) CreateSnapshot(ctx context.Context, assessment *assessmentv1alpha1.ClusterAssessment) (*assessmentv1alpha1.DeltaSummary, int, error) {
	logger := log.FromContext(ctx)

	// Convert findings to compact format
	compactFindings := compactFindings(assessment.Status.Findings)

	// Get previous snapshot for delta computation
	previousSnapshots, err := m.GetHistory(ctx, assessment.Name, 1)
	if err != nil {
		logger.Error(err, "Failed to get previous snapshots, proceeding without delta")
	}

	var delta *assessmentv1alpha1.DeltaSummary
	var previousName string
	if len(previousSnapshots) > 0 {
		prev := &previousSnapshots[0]
		previousName = prev.Name
		delta = ComputeDelta(compactFindings, assessment.Status.Summary.Score, prev)
	}

	// Create snapshot CR
	now := metav1.Now()
	snapshotName := fmt.Sprintf("%s-%s", assessment.Name, now.Format("20060102-150405"))

	snapshot := &assessmentv1alpha1.AssessmentSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: snapshotName,
			Labels: map[string]string{
				LabelAssessmentName:            assessment.Name,
				"app.kubernetes.io/managed-by": "cluster-assessment-operator",
				"app.kubernetes.io/name":       "cluster-assessment-operator",
			},
		},
		Spec: assessmentv1alpha1.AssessmentSnapshotSpec{
			AssessmentName: assessment.Name,
			Profile:        assessment.Spec.Profile,
		},
	}

	if err := m.client.Create(ctx, snapshot); err != nil {
		return nil, 0, fmt.Errorf("failed to create snapshot: %w", err)
	}

	// Update snapshot status (separate step since status is a subresource)
	snapshot.Status = assessmentv1alpha1.AssessmentSnapshotStatus{
		RunTime:              now,
		Summary:              assessment.Status.Summary,
		ClusterInfo:          assessment.Status.ClusterInfo,
		Findings:             compactFindings,
		Delta:                delta,
		PreviousSnapshotName: previousName,
	}

	if err := m.client.Status().Update(ctx, snapshot); err != nil {
		return nil, 0, fmt.Errorf("failed to update snapshot status: %w", err)
	}

	// Prune old snapshots
	historyLimit := 90
	if assessment.Spec.HistoryLimit != nil {
		historyLimit = *assessment.Spec.HistoryLimit
	}
	snapshotCount, err := m.PruneHistory(ctx, assessment.Name, historyLimit)
	if err != nil {
		logger.Error(err, "Failed to prune old snapshots")
	}

	logger.Info("Created assessment snapshot", "snapshot", snapshotName, "delta", delta != nil)
	return delta, snapshotCount, nil
}

// GetHistory returns snapshots for an assessment, sorted by runTime descending.
func (m *SnapshotManager) GetHistory(ctx context.Context, assessmentName string, limit int) ([]assessmentv1alpha1.AssessmentSnapshot, error) {
	snapshotList := &assessmentv1alpha1.AssessmentSnapshotList{}
	labelSelector := labels.SelectorFromSet(map[string]string{
		LabelAssessmentName: assessmentName,
	})

	if err := m.client.List(ctx, snapshotList, &client.ListOptions{
		LabelSelector: labelSelector,
	}); err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	// Sort by runTime descending (most recent first)
	items := snapshotList.Items
	sort.Slice(items, func(i, j int) bool {
		return items[i].Status.RunTime.After(items[j].Status.RunTime.Time)
	})

	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	return items, nil
}

// PruneHistory removes the oldest snapshots exceeding the limit.
// Returns the final snapshot count.
func (m *SnapshotManager) PruneHistory(ctx context.Context, assessmentName string, limit int) (int, error) {
	logger := log.FromContext(ctx)

	snapshotList := &assessmentv1alpha1.AssessmentSnapshotList{}
	labelSelector := labels.SelectorFromSet(map[string]string{
		LabelAssessmentName: assessmentName,
	})

	if err := m.client.List(ctx, snapshotList, &client.ListOptions{
		LabelSelector: labelSelector,
	}); err != nil {
		return 0, fmt.Errorf("failed to list snapshots for pruning: %w", err)
	}

	count := len(snapshotList.Items)
	if count <= limit {
		return count, nil
	}

	// Sort by runTime ascending (oldest first) for deletion
	items := snapshotList.Items
	sort.Slice(items, func(i, j int) bool {
		return items[i].Status.RunTime.Before(&items[j].Status.RunTime)
	})

	toDelete := count - limit
	for i := 0; i < toDelete; i++ {
		if err := m.client.Delete(ctx, &items[i]); err != nil {
			logger.Error(err, "Failed to delete old snapshot", "snapshot", items[i].Name)
			continue
		}
		logger.Info("Pruned old snapshot", "snapshot", items[i].Name)
	}

	return limit, nil
}

// compactFindings converts full findings to compact snapshots.
func compactFindings(findings []assessmentv1alpha1.Finding) []assessmentv1alpha1.FindingSnapshot {
	compact := make([]assessmentv1alpha1.FindingSnapshot, len(findings))
	for i, f := range findings {
		compact[i] = assessmentv1alpha1.FindingSnapshot{
			ID:        f.ID,
			Validator: f.Validator,
			Category:  f.Category,
			Status:    f.Status,
			Title:     f.Title,
			Resource:  f.Resource,
			Namespace: f.Namespace,
		}
	}
	return compact
}

// ComputeDelta computes the delta between current findings and a previous snapshot.
func ComputeDelta(current []assessmentv1alpha1.FindingSnapshot, currentScore *int, previous *assessmentv1alpha1.AssessmentSnapshot) *assessmentv1alpha1.DeltaSummary {
	if previous == nil {
		return nil
	}

	// Build maps: findingID -> status
	currentMap := make(map[string]assessmentv1alpha1.FindingStatus, len(current))
	for _, f := range current {
		currentMap[f.ID] = f.Status
	}

	previousMap := make(map[string]assessmentv1alpha1.FindingStatus, len(previous.Status.Findings))
	for _, f := range previous.Status.Findings {
		previousMap[f.ID] = f.Status
	}

	delta := &assessmentv1alpha1.DeltaSummary{}

	// New findings: in current but not in previous
	for id := range currentMap {
		if _, exists := previousMap[id]; !exists {
			delta.NewFindings = append(delta.NewFindings, id)
		}
	}

	// Resolved findings: in previous but not in current
	for id := range previousMap {
		if _, exists := currentMap[id]; !exists {
			delta.ResolvedFindings = append(delta.ResolvedFindings, id)
		}
	}

	// Regressions and improvements: status changed for existing findings
	for id, currentStatus := range currentMap {
		previousStatus, exists := previousMap[id]
		if !exists {
			continue
		}
		if currentStatus == previousStatus {
			continue
		}
		if severityLevel(currentStatus) > severityLevel(previousStatus) {
			delta.RegressionFindings = append(delta.RegressionFindings, id)
		} else {
			delta.ImprovedFindings = append(delta.ImprovedFindings, id)
		}
	}

	// Score delta
	if currentScore != nil && previous.Status.Summary.Score != nil {
		scoreDiff := *currentScore - *previous.Status.Summary.Score
		delta.ScoreDelta = &scoreDiff
	}

	// Sort all slices for deterministic output
	sort.Strings(delta.NewFindings)
	sort.Strings(delta.ResolvedFindings)
	sort.Strings(delta.RegressionFindings)
	sort.Strings(delta.ImprovedFindings)

	return delta
}

// severityLevel returns a numeric level for comparison.
// Higher = more severe.
func severityLevel(s assessmentv1alpha1.FindingStatus) int {
	switch s {
	case assessmentv1alpha1.FindingStatusInfo:
		return 0
	case assessmentv1alpha1.FindingStatusPass:
		return 1
	case assessmentv1alpha1.FindingStatusWarn:
		return 2
	case assessmentv1alpha1.FindingStatusFail:
		return 3
	default:
		return 0
	}
}

// SeverityLevel is exported for testing.
var SeverityLevel = severityLevel

// Ensure time import is used
var _ = time.Now
