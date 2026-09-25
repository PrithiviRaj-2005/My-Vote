package utils

import (
	"testing"
)

func TestGenerateAndValidateJWT(t *testing.T) {
	userID := "60d5ecb8b5c9c82b88f1a123"
	email := "test@pulsevote.com"
	name := "Pulse Tester"

	// 1. Generate Token
	token, err := GenerateJWT(userID, email, name)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	if token == "" {
		t.Fatalf("Expected non-empty token string")
	}

	// 2. Validate Token
	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("Failed to validate valid JWT: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}
	if claims.Name != name {
		t.Errorf("Expected Name %s, got %s", name, claims.Name)
	}

	// 3. Test Invalid Token
	_, err = ValidateJWT(token + "corrupt")
	if err == nil {
		t.Errorf("Expected error validating corrupted token, got nil")
	}
}
