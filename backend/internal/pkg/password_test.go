package pkg

import (
	"testing"
	"unicode"
)

func TestHashAndVerifyPassword(t *testing.T) {
	password := "mypassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == password {
		t.Fatal("Hash should not equal plaintext")
	}

	if !CheckPassword(password, hash) {
		t.Fatal("CheckPassword should return true for correct password")
	}
}

func TestCheckPasswordWrong(t *testing.T) {
	hash, _ := HashPassword("correct")
	if CheckPassword("wrong", hash) {
		t.Fatal("CheckPassword should return false for wrong password")
	}
}

func TestGenerateRandomPasswordLength(t *testing.T) {
	pwd := GenerateRandomPassword(16)
	if len(pwd) != 16 {
		t.Fatalf("Expected length 16, got %d", len(pwd))
	}

	pwd2 := GenerateRandomPassword(32)
	if len(pwd2) != 32 {
		t.Fatalf("Expected length 32, got %d", len(pwd2))
	}
}

func TestGenerateRandomPasswordAlphanumericOnly(t *testing.T) {
	for i := 0; i < 100; i++ {
		pwd := GenerateRandomPassword(16)
		for _, ch := range pwd {
			if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
				t.Fatalf("Password contains non-alphanumeric character %q: %s", ch, pwd)
			}
		}
	}
}

func TestGenerateRandomPasswordUniqueness(t *testing.T) {
	p1 := GenerateRandomPassword(16)
	p2 := GenerateRandomPassword(16)
	if p1 == p2 {
		t.Fatal("Two generated passwords should not be identical")
	}
}
