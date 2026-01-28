package etcdbackup

import (
	"testing"
)

func BenchmarkContainsBackupKeyword(b *testing.B) {
	testCases := []string{
		"etcd-backup-job",
		"cluster-backup-cron",
		"my-velero-job",
		"oadp-installation",
		"random-job-name",
		"nginx-deployment",
		"postgres-db",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			containsBackupKeyword(tc)
		}
	}
}

func TestContainsBackupKeyword(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"exact match", "etcd-backup", true},
		{"partial match", "my-etcd-backup-job", true},
		{"no match", "nginx", false},
		{"keyword match velero", "velero-backup", true},
		{"keyword match oadp", "oadp-operator", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsBackupKeyword(tt.input); got != tt.expected {
				t.Errorf("containsBackupKeyword(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
