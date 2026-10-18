package domain

import (
	"strings"
	"testing"
)

func TestPasswordBoundsCountUnicodeCharacters(t *testing.T) {
	for _, tt := range []struct {
		name, value string
		valid       bool
	}{
		{"short", strings.Repeat("a", 14), false}, {"minimum", strings.Repeat("ñ", 15), true},
		{"maximum", strings.Repeat("🙂", 128), true}, {"too long", strings.Repeat("a", 129), false},
		{"spaces retained", "  frase con espacios  ", true}, {"invalid utf8", strings.Repeat("a", 15) + "\xff", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if (Password([]byte(tt.value)) == nil) != tt.valid {
				t.Fatal("wrong password validation")
			}
		})
	}
}

func TestUsername(t *testing.T) {
	value, err := Username("  Isai  ")
	if err != nil || value != "Isai" {
		t.Fatal("access name trim failed")
	}
	for _, s := range []string{" ", "isai\nadmin", strings.Repeat("a", 151), "\xff"} {
		if _, err := Username(s); err == nil {
			t.Fatal("invalid access name accepted")
		}
	}
}
