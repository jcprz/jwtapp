// +build integration

package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	main "github.com/jcprz/jwtapp"
	"github.com/jcprz/jwtapp/models"
)

var a main.App

func TestMain(m *testing.M) {
	a = main.App{}
	a.Initialize()

	ensureTableExists()

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func ensureTableExists() {
	if a.DB != nil {
		a.DB.Exec("CREATE TABLE IF NOT EXISTS USERS (ID SERIAL PRIMARY KEY, EMAIL VARCHAR(50), PASSWORD VARCHAR(100));")
	}
}

func clearTable() {
	if a.DB != nil {
		a.DB.Exec("DELETE FROM users")
		a.DB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
	}
}

func TestHealthzEndpoint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/healthz", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["alive"] != true {
		t.Errorf("Expected alive to be true. Got %v", m["alive"])
	}
}

func TestSignupAPI(t *testing.T) {
	clearTable()

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid signup",
			payload:        `{"email":"test@example.com", "password":"password123"}`,
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var m map[string]interface{}
				json.Unmarshal(response.Body.Bytes(), &m)

				if m["email"] != "test@example.com" {
					t.Errorf("Expected email 'test@example.com'. Got '%v'", m["email"])
				}

				if m["password"] != nil && m["password"] != "" {
					t.Errorf("Expected password to be empty. Got '%v'", m["password"])
				}

				if m["id"] == nil {
					t.Error("Expected id to be returned")
				}
			},
		},
		{
			name:           "Missing email",
			payload:        `{"password":"password123"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var m map[string]interface{}
				json.Unmarshal(response.Body.Bytes(), &m)

				if m["message"] != "Email is missing." {
					t.Errorf("Expected error message about missing email. Got '%v'", m["message"])
				}
			},
		},
		{
			name:           "Missing password",
			payload:        `{"email":"test2@example.com"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var m map[string]interface{}
				json.Unmarshal(response.Body.Bytes(), &m)

				if m["message"] != "Password is missing." {
					t.Errorf("Expected error message about missing password. Got '%v'", m["message"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer([]byte(tt.payload)))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req)
			checkResponseCode(t, tt.expectedStatus, response.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestLoginAPI(t *testing.T) {
	clearTable()

	// First create a user
	signupPayload := `{"email":"login@example.com", "password":"password123"}`
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer([]byte(signupPayload)))
	req.Header.Set("Content-Type", "application/json")
	executeRequest(req)

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid login",
			payload:        `{"email":"login@example.com", "password":"password123"}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var m map[string]interface{}
				json.Unmarshal(response.Body.Bytes(), &m)

				if m["token"] == nil || m["token"] == "" {
					t.Error("Expected token to be returned")
				}

				// Check Authorization header
				authHeader := response.Header().Get("Authorization")
				if authHeader == "" {
					t.Error("Expected Authorization header to be set")
				}
			},
		},
		{
			name:           "Invalid password",
			payload:        `{"email":"login@example.com", "password":"wrongpassword"}`,
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var m map[string]interface{}
				json.Unmarshal(response.Body.Bytes(), &m)

				if m["message"] != "Invalid credentials." {
					t.Errorf("Expected invalid credentials message. Got '%v'", m["message"])
				}
			},
		},
		{
			name:           "Missing email",
			payload:        `{"password":"password123"}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var m map[string]interface{}
				json.Unmarshal(response.Body.Bytes(), &m)

				if m["message"] != "Email is missing." {
					t.Errorf("Expected error about missing email. Got '%v'", m["message"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(tt.payload)))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req)
			checkResponseCode(t, tt.expectedStatus, response.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestProtectedEndpoint(t *testing.T) {
	clearTable()

	// Create user and login to get token
	signupPayload := `{"email":"protected@example.com", "password":"password123"}`
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer([]byte(signupPayload)))
	req.Header.Set("Content-Type", "application/json")
	executeRequest(req)

	loginPayload := `{"email":"protected@example.com", "password":"password123"}`
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(loginPayload)))
	req.Header.Set("Content-Type", "application/json")
	loginResponse := executeRequest(req)

	var loginResult models.JWT
	json.Unmarshal(loginResponse.Body.Bytes(), &loginResult)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		description    string
	}{
		{
			name:           "Valid token",
			token:          fmt.Sprintf("Bearer %s", loginResult.Token),
			expectedStatus: http.StatusOK,
			description:    "Should allow access with valid token",
		},
		{
			name:           "Missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should reject request without token",
		},
		{
			name:           "Invalid token format",
			token:          "InvalidTokenFormat",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should reject malformed token",
		},
		{
			name:           "Invalid Bearer token",
			token:          "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			description:    "Should reject invalid JWT token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/protected", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			response := executeRequest(req)
			checkResponseCode(t, tt.expectedStatus, response.Code)
		})
	}
}

func TestDeleteUserAPI(t *testing.T) {
	clearTable()

	// Create a user to delete
	signupPayload := `{"email":"delete@example.com", "password":"password123"}`
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer([]byte(signupPayload)))
	req.Header.Set("Content-Type", "application/json")
	executeRequest(req)

	tests := []struct {
		name           string
		payload        string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid deletion",
			payload:        `{"email":"delete@example.com"}`,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				body := response.Body.String()
				if body == "" {
					t.Error("Expected response body")
				}
			},
		},
		{
			name:           "Delete non-existent user",
			payload:        `{"email":"nonexistent@example.com"}`,
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var m map[string]interface{}
				json.Unmarshal(response.Body.Bytes(), &m)

				if m["message"] != "User not found" {
					t.Errorf("Expected user not found message. Got '%v'", m["message"])
				}
			},
		},
		{
			name:           "Missing email",
			payload:        `{}`,
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, response *httptest.ResponseRecorder) {
				var m map[string]interface{}
				json.Unmarshal(response.Body.Bytes(), &m)

				if m["message"] != "Email is missing." {
					t.Errorf("Expected missing email message. Got '%v'", m["message"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("DELETE", "/delete", bytes.NewBuffer([]byte(tt.payload)))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req)
			checkResponseCode(t, tt.expectedStatus, response.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestFullUserLifecycle(t *testing.T) {
	clearTable()

	email := "lifecycle@example.com"
	password := "password123"

	// 1. Signup
	signupPayload := fmt.Sprintf(`{"email":"%s", "password":"%s"}`, email, password)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer([]byte(signupPayload)))
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	// 2. Login
	loginPayload := fmt.Sprintf(`{"email":"%s", "password":"%s"}`, email, password)
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(loginPayload)))
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	var loginResult models.JWT
	json.Unmarshal(response.Body.Bytes(), &loginResult)

	if loginResult.Token == "" {
		t.Fatal("Expected token from login")
	}

	// 3. Access protected endpoint
	req, _ = http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", loginResult.Token))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// 4. Delete user
	deletePayload := fmt.Sprintf(`{"email":"%s"}`, email)
	req, _ = http.NewRequest("DELETE", "/delete", bytes.NewBuffer([]byte(deletePayload)))
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	// 5. Verify user is deleted (login should fail)
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(loginPayload)))
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)
	// Should get error (not 200)
	if response.Code == http.StatusOK {
		t.Error("Expected login to fail after user deletion")
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	a.Router.ServeHTTP(rec, req)
	return rec
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}
