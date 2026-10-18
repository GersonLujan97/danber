// Package password implements the credential adapter. It never logs secrets.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Argon2id struct{}

func (Argon2id) Hash(password []byte) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey(password, salt, 2, 19456, 1, 32)
	defer clear(hash)
	return fmt.Sprintf("$argon2id$v=19$m=19456,t=2,p=1$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash)), nil
}

// Only the explicitly supported PHC policy is accepted. Never allocate memory
// or iterations from untrusted/arbitrary parameters in a corrupted SQL record.
func (Argon2id) Verify(encoded string, password []byte) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=19456,t=2,p=1" {
		return false, errors.New("formato de credencial no admitido")
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) != 16 {
		return false, errors.New("salt inválido")
	}
	expected, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(expected) != 32 {
		return false, errors.New("hash inválido")
	}
	actual := argon2.IDKey(password, salt, 2, 19456, 1, 32)
	defer clear(actual)
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}
