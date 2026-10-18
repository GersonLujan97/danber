package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GersonLujan97/danber/backend/internal/pacientes/internal/domain"
)

func TestDeniedWithoutStoreAccess(t *testing.T) {
	s := Service{} // A nil store detects accidental access before authorization.
	a := Actor{ID: "00000000-0000-4000-8000-000000000001"}
	if _, e := s.Register(context.Background(), a, "", domain.Registration{}); !errors.Is(e, ErrForbidden) {
		t.Fatal(e)
	}
	if _, e := s.Find(context.Background(), a, domain.Document{}); !errors.Is(e, ErrForbidden) {
		t.Fatal(e)
	}
	if _, e := s.Get(context.Background(), a, "bad"); !errors.Is(e, ErrForbidden) {
		t.Fatal(e)
	}
	if _, e := s.Types(context.Background(), a); !errors.Is(e, ErrForbidden) {
		t.Fatal(e)
	}
	// Even a forged permission with no server-issued actor is denied.
	a = Actor{Permissions: []string{"pacientes.crear", "pacientes.leer"}}
	if _, e := s.Get(context.Background(), a, "bad"); !errors.Is(e, ErrForbidden) {
		t.Fatal(e)
	}
}
func TestLimaDateAndNormalizationArePassedToStore(t *testing.T) {
	zone, _ := time.LoadLocation("America/Lima")
	store := &recordingStore{}
	s := Service{Store: store, Location: zone, Now: func() time.Time { return time.Date(2026, 10, 19, 2, 0, 0, 0, time.UTC) }}
	r := domain.Registration{Names: " Lucía ", Surnames: "Ficticia", BirthDate: "1990-01-01", Address: "Calle ficticia", Phone: "111", Document: domain.Document{Type: "dni", Number: "00123456"}}
	a := Actor{ID: "00000000-0000-4000-8000-000000000001", Permissions: []string{"pacientes.crear"}}
	key := "00000000-0000-4000-8000-000000000002"
	if _, e := s.Register(context.Background(), a, key, r); e != nil {
		t.Fatal(e)
	}
	if store.day != "2026-10-18" || store.names != "Lucía" || len(store.hash) != 32 || r.Names != " Lucía " {
		t.Fatal("wrong date, normalization, hash or caller mutation")
	}
}

type recordingStore struct {
	day, names string
	hash       []byte
}

func (s *recordingStore) Register(_ context.Context, _, _ string, h []byte, r domain.Registration, day string) (Result, error) {
	s.day = day
	s.names = r.Names
	s.hash = h
	return Result{}, nil
}
func (*recordingStore) Find(context.Context, domain.Document) (domain.Summary, error) {
	panic("unexpected")
}
func (*recordingStore) Get(context.Context, string) (domain.Patient, error)  { panic("unexpected") }
func (*recordingStore) Types(context.Context) ([]domain.DocumentType, error) { panic("unexpected") }
