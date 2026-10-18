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

func TestVerifierRejectsWrongPasswordsAndUnboundedParameters(t *testing.T) {
	hash, err := (Argon2id{}).Hash([]byte("frase ficticia de pruebas"))
	if err != nil {
		t.Fatal(err)
	}
	valid, err := (Argon2id{}).Verify(hash, []byte("frase ficticia de pruebas"))
	if err != nil || !valid {
		t.Fatal("valid credential rejected")
	}
	valid, err = (Argon2id{}).Verify(hash, []byte("otra frase ficticia"))
	if err != nil || valid {
		t.Fatal("wrong password accepted")
	}
	for _, encoded := range []string{"", strings.Replace(hash, "m=19456", "m=999999999", 1), strings.Replace(hash, "v=19", "v=16", 1), hash + "$extra"} {
		if _, err = (Argon2id{}).Verify(encoded, []byte("test")); err == nil {
			t.Fatal("unsupported PHC accepted")
		}
	}
}
