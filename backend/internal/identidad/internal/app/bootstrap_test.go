package app

import (
	"context"
	"errors"
	"testing"

	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/domain"
)

type fakeStore struct {
	calls   int
	account InitialAccount
}

func (s *fakeStore) CreateInitial(_ context.Context, a InitialAccount) (string, error) {
	s.calls++
	s.account = a
	return "synthetic-id", nil
}

type fakeHasher struct {
	calls int
	err   error
}

func (h *fakeHasher) Hash(_ []byte) (string, error) { h.calls++; return "encoded-hash", h.err }

func TestInvalidInputNeverReachesHashOrStore(t *testing.T) {
	for _, tt := range []struct{ user, password, operator string }{
		{"", "frase ficticia valida", "tester"}, {"Isai", "short", "tester"}, {"Isai", "frase ficticia valida", ""},
	} {
		s, h := &fakeStore{}, &fakeHasher{}
		_, err := (Bootstrap{s, h}).Create(context.Background(), tt.user, []byte(tt.password), tt.operator)
		if err == nil || s.calls != 0 || h.calls != 0 {
			t.Fatal("invalid input reached infrastructure")
		}
	}
}

func TestHashFailureDoesNotWriteAndSuccessfulStoreGetsNoPlaintext(t *testing.T) {
	s, h := &fakeStore{}, &fakeHasher{err: errors.New("private infrastructure detail")}
	b := Bootstrap{s, h}
	_, err := b.Create(context.Background(), "  Isai  ", []byte("frase ficticia valida"), "tester")
	if !errors.Is(err, domain.ErrPersistence) || s.calls != 0 {
		t.Fatal("hash failure leaked or wrote data")
	}
	h.err = nil
	id, err := b.Create(context.Background(), "  Isai  ", []byte("frase ficticia valida"), "tester")
	if err != nil || id != "synthetic-id" || s.account.Username != "Isai" || s.account.PasswordHash != "encoded-hash" {
		t.Fatal("invalid store input")
	}
}
