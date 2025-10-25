// +build integration

package main_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jcprz/jwtapp/database"
)

// These are legacy tests from the original implementation
// The app variable 'a' and helper functions (executeRequest, checkResponseCode, clearTable)
// are defined in integration_test.go

func TestConnectionPostgres(t *testing.T) {
	db := database.ConnectDB()
	if db == nil {
		t.Error("Failed to connect to Postgres")
	}
}

func TestConnectionRedis(t *testing.T) {
	rds := database.ConnectRedis()
	if rds == nil {
		t.Error("Failed to connect to Redis")
	}
}

func TestCreateUserAPILegacy(t *testing.T) {
	clearTable()

	var jsonStr = []byte(`{"email":"test@email.com", "password": "123456"}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["email"] != "test@email.com" {
		t.Errorf("Expected email to be 'test@email.com'. Got '%v'", m["email"])
	}

	if m["password"] != nil && m["password"] != "" {
		t.Errorf("Expected password to be empty. Got '%v'", m["password"])
	}
}

func TestLoginUserAPILegacy(t *testing.T) {
	clearTable()

	// This test was incorrectly named - it's actually testing signup
	var jsonStr = []byte(`{"email":"test@email.com", "password": "123456"}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["email"] != "test@email.com" {
		t.Errorf("Expected email to be 'test@email.com'. Got '%v'", m["email"])
	}

	if m["password"] != nil && m["password"] != "" {
		t.Errorf("Expected password to be empty. Got '%v'", m["password"])
	}
}
