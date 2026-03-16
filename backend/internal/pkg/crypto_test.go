package pkg

import (
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef" // 32 bytes for AES-256
	plaintext := "my-secret-password"

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if ciphertext == plaintext {
		t.Fatal("Ciphertext should not equal plaintext")
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("Decrypted text %q does not match original %q", decrypted, plaintext)
	}
}

func TestEncryptDifferentCiphertexts(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	plaintext := "same-text"

	c1, _ := Encrypt(plaintext, key)
	c2, _ := Encrypt(plaintext, key)

	if c1 == c2 {
		t.Fatal("Two encryptions of the same plaintext should produce different ciphertexts (random nonce)")
	}
}

func TestEncryptInvalidKeyLength(t *testing.T) {
	_, err := Encrypt("test", "short-key")
	if err == nil {
		t.Fatal("Expected error for invalid key length")
	}
}

func TestDecryptInvalidKey(t *testing.T) {
	key1 := "0123456789abcdef0123456789abcdef"
	key2 := "abcdef0123456789abcdef0123456789"

	ciphertext, _ := Encrypt("secret", key1)
	_, err := Decrypt(ciphertext, key2)
	if err == nil {
		t.Fatal("Expected error when decrypting with wrong key")
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	ciphertext, _ := Encrypt("secret", key)

	// Tamper with the ciphertext
	tampered := ciphertext[:len(ciphertext)-2] + "xx"
	_, err := Decrypt(tampered, key)
	if err == nil {
		t.Fatal("Expected error for tampered ciphertext")
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	_, err := Decrypt("not-valid-base64!!!", key)
	if err == nil {
		t.Fatal("Expected error for invalid base64")
	}
}
