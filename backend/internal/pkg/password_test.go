package pkg

import "testing"

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
