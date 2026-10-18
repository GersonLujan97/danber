package config

import (
	"strings"
	"testing"
)

func TestLoadRejectsUnsafeConfiguration(t *testing.T) {
	base := "postgresql://danber_api@localhost:26257/danber?sslmode=disable"
	tests := []struct{ name, url, addr string }{
		{"missing", "", ""},
		{"admin", "postgresql://root@localhost:26257/danber?sslmode=disable", ""},
		{"remote", "postgresql://api@example.com:26257/danber?sslmode=disable", ""},
		{"fallback", "postgresql://api@localhost:26257/danber?sslmode=prefer", ""},
		{"override_user", base + "&user=root", ""},
		{"override_host", base + "&host=example.com", ""},
		{"duplicate_tls", base + "&sslmode=require", ""},
		{"public_http", base, "0.0.0.0:8090"},
		{"missing_http_port", base, "localhost"},
		{"invalid_http_port", base, "localhost:70000"},
		{"reserved_database", "postgresql://api@localhost:26257/system?sslmode=disable", ""},
		{"secret_in_bad_url", "postgresql://api:secret-value@local%zz/danber", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(func(key string) string {
				switch key {
				case "DANBER_DATABASE_URL":
					return tt.url
				case "DANBER_HTTP_ADDR":
					return tt.addr
				default:
					return ""
				}
			})
			if err == nil {
				t.Fatal("accepted unsafe configuration")
			}
			if strings.Contains(err.Error(), "secret-value") {
				t.Fatal("configuration error leaked secret")
			}
		})
	}
}

func TestLoadLocalDefaults(t *testing.T) {
	cfg, err := Load(func(key string) string {
		if key == "DANBER_DATABASE_URL" {
			return "postgresql://danber_api@localhost:26257/danber?sslmode=disable"
		}
		return ""
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != "127.0.0.1:8090" {
		t.Fatalf("unexpected address %s", cfg.HTTPAddr)
	}
}
