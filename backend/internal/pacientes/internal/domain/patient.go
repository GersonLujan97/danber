// Package domain defines administrative patient rules without HTTP or SQL.
package domain

import (
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type Document struct {
	Type   string `json:"tipo"`
	Number string `json:"numero"`
}
type Guardian struct {
	Names        string   `json:"nombres"`
	Surnames     string   `json:"apellidos"`
	Document     Document `json:"documento"`
	Relationship string   `json:"parentesco"`
}
type Registration struct {
	Names       string    `json:"nombres"`
	Surnames    string    `json:"apellidos"`
	BirthDate   string    `json:"fecha_nacimiento"`
	CivilStatus string    `json:"estado_civil"`
	Occupation  string    `json:"ocupacion"`
	Address     string    `json:"direccion"`
	Phone       string    `json:"telefono"`
	Email       string    `json:"correo"`
	Document    Document  `json:"documento"`
	Guardian    *Guardian `json:"apoderado,omitempty"`
}
type Validation struct {
	Field string `json:"campo"`
	Code  string `json:"codigo"`
}

func (v *Validation) Error() string    { return "datos inválidos: " + v.Field + ": " + v.Code }
func Invalid(field, code string) error { return &Validation{field, code} }

var codePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,29}$`)
var dniPattern = regexp.MustCompile(`^[0-9]{8}$`)
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func UUID(s string) bool {
	return uuidPattern.MatchString(s) && s != "00000000-0000-0000-0000-000000000000"
}

func text(value *string, field string, limit int, required bool) error {
	*value = strings.TrimSpace(*value)
	if !utf8.ValidString(*value) || utf8.RuneCountInString(*value) > limit {
		return Invalid(field, "longitud_o_texto_invalido")
	}
	if required && *value == "" {
		return Invalid(field, "obligatorio")
	}
	for _, r := range *value {
		if unicode.IsControl(r) {
			return Invalid(field, "texto_invalido")
		}
	}
	return nil
}
func (d *Document) Normalize(prefix string) error {
	d.Type = strings.ToUpper(strings.TrimSpace(d.Type))
	if !codePattern.MatchString(d.Type) {
		return Invalid(prefix+".tipo", "tipo_invalido")
	}
	d.Number = strings.ToUpper(strings.TrimSpace(d.Number))
	if err := text(&d.Number, prefix+".numero", 40, true); err != nil {
		return err
	}
	if d.Type == "DNI" && !dniPattern.MatchString(d.Number) {
		return Invalid(prefix+".numero", "dni_ocho_digitos")
	}
	return nil
}

// Civil dates are evaluated in the consultorio's local day, never as elapsed hours.
func Age(birth, today time.Time) int {
	n := today.Year() - birth.Year()
	if today.Month() < birth.Month() || (today.Month() == birth.Month() && today.Day() < birth.Day()) {
		n--
	}
	return n
}
func NeedsGuardian(birth, today time.Time) bool { n := Age(birth, today); return n < 18 || n >= 66 }

func (r *Registration) Normalize() error {
	fields := []struct {
		value    *string
		field    string
		limit    int
		required bool
	}{
		{&r.Names, "nombres", 150, true}, {&r.Surnames, "apellidos", 150, true},
		{&r.CivilStatus, "estado_civil", 30, false}, {&r.Occupation, "ocupacion", 150, false},
		{&r.Address, "direccion", 500, true}, {&r.Phone, "telefono", 30, true}, {&r.Email, "correo", 254, false},
	}
	for _, f := range fields {
		if err := text(f.value, f.field, f.limit, f.required); err != nil {
			return err
		}
	}
	if r.Email != "" {
		a, e := mail.ParseAddress(r.Email)
		if e != nil || a.Address != r.Email {
			return Invalid("correo", "formato_invalido")
		}
	}
	birth, err := time.Parse("2006-01-02", r.BirthDate)
	if err != nil || birth.Year() < 1 {
		return Invalid("fecha_nacimiento", "fecha_invalida_o_futura")
	}
	if err = r.Document.Normalize("documento"); err != nil {
		return err
	}
	if r.Guardian == nil {
		return nil
	}
	g := r.Guardian
	if err = text(&g.Names, "apoderado.nombres", 150, true); err != nil {
		return err
	}
	if err = text(&g.Surnames, "apoderado.apellidos", 150, true); err != nil {
		return err
	}
	if err = text(&g.Relationship, "apoderado.parentesco", 80, true); err != nil {
		return err
	}
	if err = g.Document.Normalize("apoderado.documento"); err != nil {
		return err
	}
	if g.Document == r.Document {
		return Invalid("apoderado.documento", "debe_ser_otra_persona")
	}
	return nil
}

// ValidateDate applies only to NEW registrations. Replays return the previously
// committed result even if the patient has since had a birthday.
func (r Registration) ValidateDate(today time.Time) error {
	birth, err := time.Parse("2006-01-02", r.BirthDate)
	if err != nil || birth.Year() < 1 || r.BirthDate > today.Format("2006-01-02") {
		return Invalid("fecha_nacimiento", "fecha_invalida_o_futura")
	}
	if r.Guardian == nil && NeedsGuardian(birth, today) {
		return Invalid("apoderado", "obligatorio_por_edad")
	}
	return nil
}

type Patient struct {
	ID      string `json:"id"`
	Version int64  `json:"version"`
	State   string `json:"estado"`
	Registration
}
type Summary struct {
	ID       string   `json:"id"`
	Names    string   `json:"nombres"`
	Surnames string   `json:"apellidos"`
	State    string   `json:"estado"`
	Document Document `json:"documento"`
}
type DocumentType struct {
	Code string `json:"codigo"`
	Name string `json:"nombre"`
}
