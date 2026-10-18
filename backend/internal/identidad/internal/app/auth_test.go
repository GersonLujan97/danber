package app

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type authStore struct {
	credential     Credential
	hash, previous []byte
	started        int
	err            error
}

func (s *authStore) FindCredential(context.Context, string) (Credential, error) {
	return s.credential, s.err
}
func (s *authStore) StartSession(_ context.Context, _ Credential, hash, previous []byte, _ time.Duration) (Session, error) {
	s.hash = hash
	s.previous = previous
	s.started++
	return Session{Principal: Principal{ID: "synthetic"}, Expires: time.Now().Add(time.Hour)}, s.err
}
func (s *authStore) TouchSession(context.Context, []byte, time.Duration) (Session, error) {
	return Session{}, s.err
}
func (s *authStore) Revoke(context.Context, []byte, time.Duration, bool) error { return s.err }

type verifier struct {
	seen  string
	valid bool
}

func (v *verifier) Verify(hash string, _ []byte) (bool, error) { v.seen = hash; return v.valid, nil }
func policy() Policy                                           { return Policy{Idle: 15 * time.Minute, Absolute: 8 * time.Hour} }

func TestUnknownWrongAndSuspendedUseSameFailureAndHashWork(t *testing.T) {
	for _, c := range []Credential{{}, {ID: "user", Hash: "real", Active: false}, {ID: "user", Hash: "real", Active: true}} {
		store := &authStore{credential: c}
		verify := &verifier{valid: c.ID == "" || !c.Active}
		a, _ := NewAuth(store, verify, "dummy", policy())
		_, err := a.Login(context.Background(), "Isai", []byte("incorrecta"), "")
		if !errors.Is(err, ErrCredentials) || store.started != 0 || verify.seen == "" {
			t.Fatal("failure behavior differs or verification skipped")
		}
		if c.ID == "" && verify.seen != "dummy" {
			t.Fatal("unknown account must perform dummy verification")
		}
	}
}
func TestSessionTokenIsRandomAndOnlyItsHashReachesSQL(t *testing.T) {
	s := &authStore{credential: Credential{ID: "user", Hash: "encoded", Active: true}}
	a, _ := NewAuth(s, &verifier{valid: true}, "dummy", policy())
	first, err := a.Login(context.Background(), "Isai", []byte("frase ficticia de pruebas"), "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := a.Login(context.Background(), "Isai", []byte("frase ficticia de pruebas"), first.Token)
	if err != nil {
		t.Fatal(err)
	}
	if first.Token == second.Token || len(s.hash) != 32 || !bytes.Equal(s.hash, TokenHash(second.Token)) || !bytes.Equal(s.previous, TokenHash(first.Token)) {
		t.Fatal("token/hash/rotation contract failed")
	}
	if bytes.Contains(s.hash, []byte(second.Token)) {
		t.Fatal("raw token reached storage")
	}
}

type blockingVerifier struct {
	started chan struct{}
	release chan struct{}
}

func (v blockingVerifier) Verify(string, []byte) (bool, error) {
	v.started <- struct{}{}
	<-v.release
	return false, nil
}
func TestExpensiveVerificationHasBoundedConcurrency(t *testing.T) {
	v := blockingVerifier{make(chan struct{}, 2), make(chan struct{})}
	a, _ := NewAuth(&authStore{credential: Credential{ID: "user", Hash: "encoded", Active: true}}, v, "dummy", policy())
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); _, _ = a.Login(context.Background(), "Isai", []byte("incorrecta"), "") }()
	}
	<-v.started
	<-v.started
	_, err := a.Login(context.Background(), "Isai", []byte("incorrecta"), "")
	close(v.release)
	wg.Wait()
	if !errors.Is(err, ErrBusy) {
		t.Fatal("third expensive operation should be refused without waiting")
	}
}
func TestPermissionContractDeniesUnassignedAndEmptyPermissions(t *testing.T) {
	p := Principal{Roles: []string{"maestro"}, Permissions: []string{"pacientes.registrar"}}
	if !p.Allows("pacientes.registrar") || p.Allows("economia.cobrar") || p.Allows("") {
		t.Fatal("permission bypass")
	}
}
