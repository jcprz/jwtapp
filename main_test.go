// +build integration

package main_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	main "github.com/jcprz/jwtapp"
	"github.com/jcprz/jwtapp/database"
)

var a main.App

func TestMain(m *testing.M) {
	a = main.App{}
	a.Initialize()

	code := m.Run()

	clearTable()

	os.Exit(code)
}
func TestConnectionPostgres(t *testing.T) {
	a.DB = database.ConnectDB()

}

func TestConnectionRedis(t *testing.T) {
	a.RDS = database.ConnectRedis()

}

func TestCreateUserAPI(t *testing.T) {
	// clearTable()
	var jsonStr = []byte(`{"email":"test@email.com", "password": "123456"}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["email"] != "test@email.com" {
		t.Errorf("Expected email name to be 'test@email.com'. Got '%v'", m["name"])
	}

	if m["password"] != "" {
		t.Errorf("Expected password to retunr empty. Got '%v'", m["price"])
	}

}

func TestLoginUserAPI(t *testing.T) {
	var jsonStr = []byte(`{"email":"test@email.com", "password": "123456"}`)
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["email"] != "test@email.com" {
		t.Errorf("Expected email name to be 'test@email.com'. Got '%v'", m["name"])
	}

	if m["password"] != "" {
		t.Errorf("Expected password to retunr empty. Got '%v'", m["price"])
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	log.Printf("Request URI: %s | Request Body: %s | Request Header: %s", req.URL, req.Body, req.Header)
	rec := httptest.NewRecorder()
	a.Router.ServeHTTP(rec, req)
	// http.NewServeMux().ServeHTTP(rec, req)

	return rec
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM users")
	a.DB.Exec("ALTER SEQUENCE users RESTART WITH 1")
}
