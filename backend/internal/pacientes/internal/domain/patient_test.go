package domain

import (
	"strings"
	"testing"
	"time"
)

func registration() Registration {
	return Registration{Names: " Lucía ", Surnames: "O'Álvarez", BirthDate: "1990-01-01", Address: "Calle ficticia 1", Phone: "+51 (01) 111-1111", Document: Document{"dni", "00123456"}}
}
func TestAgeBoundariesAndLeapBirthday(t *testing.T) {
	for _, c := range []struct {
		birth, day string
		needs      bool
	}{
		{"2008-10-19", "2026-10-18", true}, {"2008-10-18", "2026-10-18", false},
		{"1960-10-19", "2026-10-18", false}, {"1960-10-18", "2026-10-18", true},
		{"2008-02-29", "2026-02-28", true}, {"2008-02-29", "2026-03-01", false},
	} {
		b, _ := time.Parse("2006-01-02", c.birth)
		d, _ := time.Parse("2006-01-02", c.day)
		if NeedsGuardian(b, d) != c.needs {
			t.Fatalf("boundary %s %s", c.birth, c.day)
		}
	}
}
func TestRegistrationValidation(t *testing.T) {
	day, _ := time.Parse("2006-01-02", "2026-10-18")
	for _, c := range []struct {
		name  string
		edit  func(*Registration)
		field string
	}{
		{"names required", func(r *Registration) { r.Names = "  " }, "nombres"},
		{"birth future", func(r *Registration) { r.BirthDate = "2026-10-19" }, "fecha_nacimiento"},
		{"invalid birth", func(r *Registration) { r.BirthDate = "2026-02-29" }, "fecha_nacimiento"},
		{"phone required", func(r *Registration) { r.Phone = "" }, "telefono"},
		{"address required", func(r *Registration) { r.Address = "" }, "direccion"},
		{"email", func(r *Registration) { r.Email = "Lucía <x@example.com>" }, "correo"},
		{"dni", func(r *Registration) { r.Document.Number = "1234567" }, "documento.numero"},
		{"own doc", func(r *Registration) { r.Document.Number = "" }, "documento.numero"},
		{"minor", func(r *Registration) { r.BirthDate = "2015-01-01" }, "apoderado"},
		{"elder", func(r *Registration) { r.BirthDate = "1960-10-18" }, "apoderado"},
		{"unicode limit", func(r *Registration) { r.Names = strings.Repeat("á", 151) }, "nombres"},
		{"control", func(r *Registration) { r.Surnames = "Un\x00nombre" }, "apellidos"},
		{"guardian relationship", func(r *Registration) {
			r.Guardian = &Guardian{Names: "María", Surnames: "Ficticia", Document: Document{"DNI", "00234567"}}
		}, "apoderado.parentesco"},
		{"guardian self", func(r *Registration) {
			r.Guardian = &Guardian{Names: "María", Surnames: "Ficticia", Document: Document{"DNI", "00123456"}, Relationship: "Madre"}
		}, "apoderado.documento"},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := registration()
			c.edit(&r)
			e := r.Normalize()
			if e == nil {
				e = r.ValidateDate(day)
			}
			v, ok := e.(*Validation)
			if !ok || v.Field != c.field {
				t.Fatalf("error=%v", e)
			}
		})
	}
	r := registration()
	if e := r.Normalize(); e != nil {
		t.Fatal(e)
	}
	if r.Names != "Lucía" || r.Document.Number != "00123456" || r.Document.Type != "DNI" {
		t.Fatal("normalization corrupted presentation or leading zeros")
	}
}
