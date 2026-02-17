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

package oadpbackup

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/profiles"
	"github.com/openshift-assessment/cluster-assessment-operator/pkg/validator"
)

const (
	validatorName        = "oadpbackup"
	validatorDescription = "Validates OADP/Velero backup schedules and last successful backups"
	validatorCategory    = "Platform"
)

func init() {
	_ = validator.Register(&OADPBackupValidator{})
}

// OADPBackupValidator checks OADP/Velero backup configuration.
type OADPBackupValidator struct{}

func (v *OADPBackupValidator) Name() string        { return validatorName }
func (v *OADPBackupValidator) Description() string { return validatorDescription }
func (v *OADPBackupValidator) Category() string    { return validatorCategory }

// Validate checks for OADP/Velero backup schedules and recent backups.
func (v *OADPBackupValidator) Validate(ctx context.Context, c client.Client, profile profiles.Profile) ([]assessmentv1alpha1.Finding, error) {
	var findings []assessmentv1alpha1.Finding

	// Check Velero Schedules
	findings = append(findings, v.checkSchedules(ctx, c)...)

	// Check recent backups
	findings = append(findings, v.checkRecentBackups(ctx, c)...)

	// If no backup-related findings, add a general check
	if len(findings) == 0 {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "oadpbackup-no-schedules",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         assessmentv1alpha1.FindingStatusWarn,
			Title:          "No Velero Backup Schedules Found",
			Description:    "No Velero Backup Schedules were detected. Regular backups are essential for disaster recovery.",
			Impact:         "Without scheduled backups, data recovery after failures may not be possible.",
			Recommendation: "Install OADP and configure Velero backup schedules for critical namespaces.",
			References: []string{
				"https://docs.openshift.com/container-platform/latest/backup_and_restore/application_backup_and_restore/oadp-intro.html",
			},
			Remediation: &assessmentv1alpha1.RemediationGuidance{
				Safety: assessmentv1alpha1.RemediationSafeApply,
				Commands: []assessmentv1alpha1.RemediationCommand{
					{Command: "oc get csv -A | grep oadp", Description: "Check if OADP operator is installed"},
					{Command: "oc get schedules -A", Description: "List all Velero backup schedules"},
				},
				DocumentationURL: "https://docs.openshift.com/container-platform/latest/backup_and_restore/application_backup_and_restore/oadp-intro.html",
				EstimatedImpact:  "Installing OADP enables automated backup management",
			},
		})
	}

	return findings, nil
}

func (v *OADPBackupValidator) checkSchedules(ctx context.Context, c client.Client) []assessmentv1alpha1.Finding {
	var findings []assessmentv1alpha1.Finding

	scheduleList := &unstructured.UnstructuredList{}
	scheduleList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "velero.io",
		Version: "v1",
		Kind:    "ScheduleList",
	})

	if err := c.List(ctx, scheduleList); err != nil {
		// Velero not installed
		return nil
	}

	if len(scheduleList.Items) == 0 {
		return nil // Will be caught by parent function
	}

	paused := 0
	active := 0
	for _, sched := range scheduleList.Items {
		name := sched.GetName()
		ns := sched.GetNamespace()
		p, _, _ := unstructured.NestedBool(sched.Object, "spec", "paused")
		schedule, _, _ := unstructured.NestedString(sched.Object, "spec", "schedule")

		if p {
			paused++
			findings = append(findings, assessmentv1alpha1.Finding{
				ID:             fmt.Sprintf("oadpbackup-schedule-paused-%s-%s", ns, name),
				Validator:      validatorName,
				Category:       validatorCategory,
				Status:         assessmentv1alpha1.FindingStatusWarn,
				Title:          fmt.Sprintf("Backup Schedule Paused: %s/%s", ns, name),
				Description:    fmt.Sprintf("Velero backup schedule '%s/%s' (cron: %s) is paused and will not create backups.", ns, name, schedule),
				Recommendation: "Review if this schedule should be unpaused.",
				Remediation: &assessmentv1alpha1.RemediationGuidance{
					Safety: assessmentv1alpha1.RemediationSafeApply,
					Commands: []assessmentv1alpha1.RemediationCommand{
						{Command: fmt.Sprintf("oc patch schedule %s -n %s --type=merge -p '{\"spec\":{\"paused\":false}}'", name, ns), Description: "Unpause the backup schedule", RequiresConfirmation: true},
					},
				},
			})
		} else {
			active++
		}
	}

	if active > 0 {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:          "oadpbackup-schedules-active",
			Validator:   validatorName,
			Category:    validatorCategory,
			Status:      assessmentv1alpha1.FindingStatusPass,
			Title:       "Active Backup Schedules Found",
			Description: fmt.Sprintf("%d active Velero backup schedule(s) found.", active),
		})
	}

	return findings
}

func (v *OADPBackupValidator) checkRecentBackups(ctx context.Context, c client.Client) []assessmentv1alpha1.Finding {
	var findings []assessmentv1alpha1.Finding

	backupList := &unstructured.UnstructuredList{}
	backupList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "velero.io",
		Version: "v1",
		Kind:    "BackupList",
	})

	if err := c.List(ctx, backupList); err != nil {
		return nil
	}

	if len(backupList.Items) == 0 {
		return nil
	}

	// Find the most recent completed backup
	var latestCompletionTime time.Time
	var latestBackupName string
	failedBackups := 0

	for _, backup := range backupList.Items {
		phase, _, _ := unstructured.NestedString(backup.Object, "status", "phase")

		if phase == "Failed" || phase == "PartiallyFailed" {
			failedBackups++
			continue
		}

		if phase != "Completed" {
			continue
		}

		completionStr, found, _ := unstructured.NestedString(backup.Object, "status", "completionTimestamp")
		if !found || completionStr == "" {
			continue
		}

		t, err := time.Parse(time.RFC3339, completionStr)
		if err != nil {
			continue
		}

		if t.After(latestCompletionTime) {
			latestCompletionTime = t
			latestBackupName = fmt.Sprintf("%s/%s", backup.GetNamespace(), backup.GetName())
		}
	}

	// Check if recent backup is stale
	if !latestCompletionTime.IsZero() {
		age := time.Since(latestCompletionTime)
		maxAge := 7 * 24 * time.Hour // 7 days default

		if age > maxAge {
			findings = append(findings, assessmentv1alpha1.Finding{
				ID:             "oadpbackup-stale-backup",
				Validator:      validatorName,
				Category:       validatorCategory,
				Status:         assessmentv1alpha1.FindingStatusWarn,
				Title:          "Last Successful Backup is Stale",
				Description:    fmt.Sprintf("Most recent successful backup (%s) completed %s ago, exceeding the 7-day threshold.", latestBackupName, formatDuration(age)),
				Impact:         "Stale backups provide inadequate protection against data loss.",
				Recommendation: "Investigate why recent backup schedules did not produce successful backups.",
			})
		} else {
			findings = append(findings, assessmentv1alpha1.Finding{
				ID:          "oadpbackup-recent-backup-ok",
				Validator:   validatorName,
				Category:    validatorCategory,
				Status:      assessmentv1alpha1.FindingStatusPass,
				Title:       "Recent Backup Available",
				Description: fmt.Sprintf("Most recent successful backup (%s) completed %s ago.", latestBackupName, formatDuration(age)),
			})
		}
	}

	if failedBackups > 0 {
		findings = append(findings, assessmentv1alpha1.Finding{
			ID:             "oadpbackup-failed-backups",
			Validator:      validatorName,
			Category:       validatorCategory,
			Status:         assessmentv1alpha1.FindingStatusWarn,
			Title:          "Failed Backups Detected",
			Description:    fmt.Sprintf("%d backup(s) have Failed or PartiallyFailed status.", failedBackups),
			Impact:         "Failed backups may indicate storage issues or misconfigured backup resources.",
			Recommendation: "Review failed backup logs and ensure backup storage is accessible.",
			Remediation: &assessmentv1alpha1.RemediationGuidance{
				Safety: assessmentv1alpha1.RemediationSafeApply,
				Commands: []assessmentv1alpha1.RemediationCommand{
					{Command: "oc get backups -A -o json | jq '.items[] | select(.status.phase==\"Failed\" or .status.phase==\"PartiallyFailed\") | .metadata.namespace + \"/\" + .metadata.name + \" (\" + .status.phase + \")\"'", Description: "List failed backups"},
					{Command: "velero backup describe <backup-name> -n <namespace>", Description: "Get details of a failed backup"},
				},
			},
		})
	}

	return findings
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	if hours < 24 {
		return fmt.Sprintf("%d hours", hours)
	}
	days := hours / 24
	return fmt.Sprintf("%d days", days)
}
