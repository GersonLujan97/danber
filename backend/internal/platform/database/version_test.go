package database

import "testing"

func TestSupportedVersion(t *testing.T) {
	for _, tt := range []struct {
		version string
		want    bool
	}{
		{"CockroachDB CCL v26.2.0 (windows amd64)", true},
		{"CockroachDB CCL v26.2.7 (linux amd64)", true},
		{"CockroachDB CCL v26.2.8 (linux amd64)", false},
		{"CockroachDB CCL v26.2.0-beta (linux amd64)", false},
		{"CockroachDB CCL v26.2.70 (linux amd64)", false},
		{"PostgreSQL 26.2.0", false},
		{"not CockroachDB v26.2.0", false},
		{"", false},
	} {
		t.Run(tt.version, func(t *testing.T) {
			if got := supportedVersion(tt.version); got != tt.want {
				t.Fatalf("supportedVersion()=%v, want %v", got, tt.want)
			}
		})
	}
}
