// Package domain contains identity rules without SQL or transport dependencies.
package domain

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrUsername      = errors.New("el usuario debe tener entre 1 y 150 caracteres, sin controles")
	ErrPassword      = errors.New("la contraseña debe tener entre 15 y 128 caracteres Unicode válidos")
	ErrOperator      = errors.New("no se pudo identificar al operador local")
	ErrInitialized   = errors.New("la cuenta Maestro inicial ya fue creada; no se modifica ninguna cuenta")
	ErrUsernameTaken = errors.New("el nombre de acceso ya existe; no se eleva ni reemplaza esa cuenta")
	ErrSchema        = errors.New("falta la migración 009 o su fila de control; aplicar las migraciones antes de continuar")
	ErrRoles         = errors.New("faltan los roles Maestro o Recepción; revisar los catálogos")
	ErrPersistence   = errors.New("no se pudo confirmar la creación; revisar la base y volver a ejecutar el comando sin modificar cuentas manualmente")
)

func Username(value string) (string, error) {
	value = strings.TrimSpace(value)
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) < 1 || utf8.RuneCountInString(value) > 150 {
		return "", ErrUsername
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return "", ErrUsername
		}
	}
	return value, nil
}

func Password(value []byte) error {
	if !utf8.Valid(value) || utf8.RuneCount(value) < 15 || utf8.RuneCount(value) > 128 {
		return ErrPassword
	}
	return nil
}

func Operator(value string) error {
	if !utf8.ValidString(value) || strings.TrimSpace(value) == "" || utf8.RuneCountInString(value) > 200 {
		return ErrOperator
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return ErrOperator
		}
	}
	return nil
}
