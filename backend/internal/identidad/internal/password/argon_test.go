package password

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"
)

func TestHashUsesSaltAndCanBeVerified(t *testing.T) {
	secret := []byte("frase ficticia para pruebas")
	first, err := (Argon2id{}).Hash(secret)
	if err != nil {
		t.Fatal("hash failed")
	}
	second, err := (Argon2id{}).Hash(secret)
	if err != nil {
		t.Fatal("hash failed")
	}
	if first == second || strings.Contains(first, string(secret)) {
		t.Fatal("salt or secret handling failed")
	}
	parts := strings.Split(first, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=19456,t=2,p=1" {
		t.Fatal("unexpected PHC format")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != 16 {
		t.Fatal("invalid salt")
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(hash) != 32 {
		t.Fatal("invalid hash")
	}
	if !bytes.Equal(hash, argon2.IDKey(secret, salt, 2, 19456, 1, 32)) {
		t.Fatal("stored hash cannot verify password")
	}
	if bytes.Equal(hash, argon2.IDKey([]byte("otra frase ficticia"), salt, 2, 19456, 1, 32)) {
		t.Fatal("different password accepted")
	}
}
