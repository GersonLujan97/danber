package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/domain"
)

var (
	ErrCredentials = errors.New("credenciales inválidas")
	ErrSession     = errors.New("sesión no válida")
	ErrUnavailable = errors.New("autenticación no disponible")
	ErrBusy        = errors.New("demasiados intentos")
)

type Policy struct{ Idle, Absolute time.Duration }

func (p Policy) Valid() bool {
	return p.Idle >= time.Minute && p.Idle <= time.Hour && p.Absolute >= 30*time.Minute && p.Absolute <= 24*time.Hour && p.Idle <= p.Absolute
}

type Principal struct {
	ID          string   `json:"id"`
	Username    string   `json:"usuario"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permisos"`
}

func (p Principal) Allows(permission string) bool {
	for _, value := range p.Permissions {
		if permission != "" && value == permission {
			return true
		}
	}
	return false
}

type Credential struct {
	ID, Hash string
	Active   bool
}
type Session struct {
	Principal Principal
	Expires   time.Time
	Token     string
}
type AuthStore interface {
	FindCredential(context.Context, string) (Credential, error)
	StartSession(context.Context, Credential, []byte, []byte, time.Duration) (Session, error)
	TouchSession(context.Context, []byte, time.Duration) (Session, error)
	Revoke(context.Context, []byte, time.Duration, bool) error
}
type Verifier interface {
	Verify(string, []byte) (bool, error)
}
type Auth struct {
	store    AuthStore
	verifier Verifier
	dummy    string
	policy   Policy
	gate     chan struct{}
}

func NewAuth(store AuthStore, verifier Verifier, dummy string, policy Policy) (*Auth, error) {
	if !policy.Valid() {
		return nil, errors.New("política de sesión inválida")
	}
	return &Auth{store: store, verifier: verifier, dummy: dummy, policy: policy, gate: make(chan struct{}, 2)}, nil
}

func (a *Auth) Login(ctx context.Context, username string, password []byte, previous string) (Session, error) {
	name, err := domain.Username(username)
	if err != nil || !utf8.Valid(password) || len(password) == 0 || utf8.RuneCount(password) > 128 {
		return Session{}, ErrCredentials
	}
	select {
	case a.gate <- struct{}{}:
		defer func() { <-a.gate }()
	default:
		return Session{}, ErrBusy
	}
	credential, err := a.store.FindCredential(ctx, name)
	if err != nil {
		return Session{}, ErrUnavailable
	}
	encoded := credential.Hash
	if credential.ID == "" {
		encoded = a.dummy
	}
	valid, err := a.verifier.Verify(encoded, password)
	if err != nil {
		_, _ = a.verifier.Verify(a.dummy, password)
		return Session{}, ErrUnavailable
	}
	if !valid || !credential.Active || credential.ID == "" {
		return Session{}, ErrCredentials
	}
	random := make([]byte, 32)
	if _, err = rand.Read(random); err != nil {
		return Session{}, ErrUnavailable
	}
	token := base64.RawURLEncoding.EncodeToString(random)
	clear(random)
	hash := TokenHash(token)
	result, err := a.store.StartSession(ctx, credential, hash, TokenHash(previous), a.policy.Absolute)
	if err != nil {
		return Session{}, err
	}
	result.Token = token
	return result, nil
}

func (a *Auth) Current(ctx context.Context, token string) (Session, error) {
	hash := TokenHash(token)
	if hash == nil {
		return Session{}, ErrSession
	}
	return a.store.TouchSession(ctx, hash, a.policy.Idle)
}

func (a *Auth) Logout(ctx context.Context, token string, all bool) error {
	hash := TokenHash(token)
	if hash == nil {
		if all {
			return ErrSession
		}
		return nil
	}
	return a.store.Revoke(ctx, hash, a.policy.Idle, all)
}

func TokenHash(token string) []byte {
	if len(token) != 43 {
		return nil
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || len(decoded) != 32 {
		return nil
	}
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
