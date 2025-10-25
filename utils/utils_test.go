package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jcprz/jwtapp/models"
	"golang.org/x/crypto/bcrypt"
)

func TestResponseJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}

	ResponseJSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestRespondWithError(t *testing.T) {
	w := httptest.NewRecorder()
	message := "Test error message"

	RespondWithError(w, http.StatusBadRequest, message)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestGenerateToken(t *testing.T) {
	// Set up environment variable for testing
	os.Setenv("SECRET", "test-secret-key")
	defer os.Unsetenv("SECRET")

	user := models.User{
		ID:    1,
		Email: "test@example.com",
	}

	token, err := GenerateToken(user)

	if err != nil {
		t.Errorf("GenerateToken() returned error: %v", err)
	}

	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}

	// Verify token can be parsed
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})

	if err != nil {
		t.Errorf("Failed to parse generated token: %v", err)
	}

	if !parsedToken.Valid {
		t.Error("Generated token is not valid")
	}

	// Check claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Error("Failed to get claims from token")
	}

	if claims["email"] != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, claims["email"])
	}

	if claims["iss"] != "course" {
		t.Errorf("Expected issuer 'course', got %s", claims["iss"])
	}

	// Check expiration exists
	if _, ok := claims["exp"]; !ok {
		t.Error("Token missing expiration claim")
	}

	// Check issued at exists
	if _, ok := claims["iat"]; !ok {
		t.Error("Token missing issued at claim")
	}
}

func TestGenerateTokenExpiration(t *testing.T) {
	os.Setenv("SECRET", "test-secret-key")
	defer os.Unsetenv("SECRET")

	user := models.User{
		ID:    1,
		Email: "test@example.com",
	}

	token, err := GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() returned error: %v", err)
	}

	parsedToken, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})

	claims := parsedToken.Claims.(jwt.MapClaims)
	exp := int64(claims["exp"].(float64))
	iat := int64(claims["iat"].(float64))

	expectedDuration := int64(24 * time.Hour / time.Second)
	actualDuration := exp - iat

	// Allow 5 second tolerance
	if actualDuration < expectedDuration-5 || actualDuration > expectedDuration+5 {
		t.Errorf("Expected token duration ~%d seconds, got %d", expectedDuration, actualDuration)
	}
}

func TestComparePasswords(t *testing.T) {
	password := "testPassword123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name           string
		hashedPassword string
		password       []byte
		expected       bool
	}{
		{
			name:           "Valid password",
			hashedPassword: string(hashedPassword),
			password:       []byte(password),
			expected:       true,
		},
		{
			name:           "Invalid password",
			hashedPassword: string(hashedPassword),
			password:       []byte("wrongPassword"),
			expected:       false,
		},
		{
			name:           "Empty password",
			hashedPassword: string(hashedPassword),
			password:       []byte(""),
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComparePasswords(tt.hashedPassword, tt.password)
			if result != tt.expected {
				t.Errorf("ComparePasswords() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
