package services

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHashing(t *testing.T) {
	password := "Secret123!"

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Comparing correct password should succeed
	err = bcrypt.CompareHashAndPassword(hashed, []byte(password))
	if err != nil {
		t.Errorf("Password comparison failed for correct password: %v", err)
	}

	// Comparing wrong password should fail
	err = bcrypt.CompareHashAndPassword(hashed, []byte("WrongPassword"))
	if err == nil {
		t.Errorf("Expected failure for incorrect password, got nil")
	}
}

func TestShareCodeGeneration(t *testing.T) {
	code1, err := generateShareCode()
	if err != nil {
		t.Fatalf("Failed to generate share code 1: %v", err)
	}

	code2, err := generateShareCode()
	if err != nil {
		t.Fatalf("Failed to generate share code 2: %v", err)
	}

	if len(code1) != 6 || len(code2) != 6 {
		t.Errorf("Expected 6 characters, got %d and %d", len(code1), len(code2))
	}

	if code1 == code2 {
		t.Errorf("Expected distinct random share codes, got duplicate: %s", code1)
	}
}

func TestPollOptionValidationLogic(t *testing.T) {
	// Rule: Min 2 options, Max 10 options, no duplicates
	options := []string{"Python", "Go", "python "}
	seen := make(map[string]bool)
	var hasDuplicate bool

	for _, opt := range options {
		lower := strings.ToLower(strings.TrimSpace(opt))
		if seen[lower] {
			hasDuplicate = true
			break
		}
		seen[lower] = true
	}

	if !hasDuplicate {
		t.Errorf("Expected duplicate check to detect 'Python' and 'python '")
	}
}
