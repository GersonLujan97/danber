package app

import (
	"context"

	"github.com/GersonLujan97/danber/backend/internal/identidad/internal/domain"
)

type InitialAccount struct{ Username, PasswordHash, Operator string }
type Store interface {
	CreateInitial(context.Context, InitialAccount) (string, error)
}
type Hasher interface{ Hash([]byte) (string, error) }

type Bootstrap struct {
	Store  Store
	Hasher Hasher
}

func (b Bootstrap) Create(ctx context.Context, username string, password []byte, operator string) (string, error) {
	name, err := domain.Username(username)
	if err != nil {
		return "", err
	}
	if err = domain.Password(password); err != nil {
		return "", err
	}
	if err = domain.Operator(operator); err != nil {
		return "", err
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	hash, err := b.Hasher.Hash(password)
	if err != nil {
		return "", domain.ErrPersistence
	}
	return b.Store.CreateInitial(ctx, InitialAccount{Username: name, PasswordHash: hash, Operator: operator})
}
