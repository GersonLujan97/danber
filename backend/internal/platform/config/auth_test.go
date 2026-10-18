package config

import (
	"testing"
	"time"
)

func TestSessionConfigurationDefaultsAndBounds(t *testing.T) {
	base := map[string]string{"DANBER_DATABASE_URL": "postgresql://danber_api@localhost:26257/danber?sslmode=disable"}
	load := func(extra map[string]string) (Config, error) {
		return Load(func(k string) string {
			if v, ok := extra[k]; ok {
				return v
			}
			return base[k]
		})
	}
	c, err := load(nil)
	if err != nil || c.AllowInsecureAuth || c.SessionIdle != 15*time.Minute || c.SessionTTL != 8*time.Hour {
		t.Fatal("unsafe defaults")
	}
	for _, extra := range []map[string]string{{"DANBER_ALLOW_INSECURE_AUTH": "1"}, {"DANBER_SESSION_IDLE": "SECRET"}, {"DANBER_SESSION_IDLE": "0s"}, {"DANBER_SESSION_TTL": "25h"}, {"DANBER_SESSION_IDLE": "1h", "DANBER_SESSION_TTL": "30m"}} {
		if _, err := load(extra); err == nil {
			t.Fatal("invalid session policy accepted")
		}
	}
	c, err = load(map[string]string{"DANBER_ALLOW_INSECURE_AUTH": "true", "DANBER_SESSION_IDLE": "10m", "DANBER_SESSION_TTL": "4h"})
	if err != nil || !c.AllowInsecureAuth || c.SessionIdle != 10*time.Minute || c.SessionTTL != 4*time.Hour {
		t.Fatal("valid local policy rejected")
	}
}
