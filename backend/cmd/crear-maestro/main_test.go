package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func env(key string) string {
	if key == "DANBER_DATABASE_URL" {
		return "postgresql://danber_api@localhost:26257/danber?sslmode=disable"
	}
	return ""
}

func TestNoPasswordArgumentsOrPipedInput(t *testing.T) {
	for _, tt := range []struct {
		args     []string
		terminal bool
	}{
		{[]string{"--usuario", "Isai", "--password", "SECRET_ARG"}, true},
		{[]string{"--usuario", "Isai", "SECRET_ARG"}, true},
		{[]string{"--usuario", "Isai"}, false},
	} {
		var out bytes.Buffer
		err := run(context.Background(), tt.args, &out, env, tt.terminal, func() ([]byte, error) { t.Fatal("secret read unexpectedly"); return nil, nil }, func(context.Context, string, []byte, string) error { t.Fatal("unexpected SQL operation"); return nil })
		if err == nil || strings.Contains(err.Error()+out.String(), "SECRET_ARG") {
			t.Fatal("unsafe arguments/input accepted or leaked")
		}
	}
}

func TestMismatchDoesNotWriteAndClearsBothBuffers(t *testing.T) {
	first, second := []byte("primera frase ficticia"), []byte("segunda frase ficticia")
	reads := 0
	var out bytes.Buffer
	err := run(context.Background(), []string{"--usuario", "Isai"}, &out, env, true, func() ([]byte, error) {
		reads++
		if reads == 1 {
			return first, nil
		}
		return second, nil
	}, func(context.Context, string, []byte, string) error { t.Fatal("write on mismatch"); return nil })
	if err == nil || bytes.Contains(out.Bytes(), []byte("frase ficticia")) {
		t.Fatal("mismatch handling failed")
	}
	if !allZero(first) || !allZero(second) {
		t.Fatal("password buffers not cleared")
	}
}

func TestSuccessfulPromptDoesNotPrintSecret(t *testing.T) {
	secret := []byte("frase ficticia valida")
	confirmation := append([]byte(nil), secret...)
	reads, calls := 0, 0
	var out bytes.Buffer
	err := run(context.Background(), []string{"--usuario", "Isai"}, &out, env, true, func() ([]byte, error) {
		reads++
		if reads == 1 {
			return secret, nil
		}
		return confirmation, nil
	}, func(_ context.Context, name string, p []byte, url string) error {
		calls++
		if name != "Isai" || string(p) != "frase ficticia valida" || !strings.Contains(url, "danber_api") {
			return errors.New("incorrect input")
		}
		return nil
	})
	if err != nil || calls != 1 || strings.Contains(out.String(), "frase ficticia valida") {
		t.Fatal("unexpected result or password output")
	}
	if !allZero(secret) || !allZero(confirmation) {
		t.Fatal("password buffers not cleared")
	}
}

func allZero(value []byte) bool {
	for _, b := range value {
		if b != 0 {
			return false
		}
	}
	return true
}
