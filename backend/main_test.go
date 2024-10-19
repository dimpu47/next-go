package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGetUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	debugMode := os.Getenv("DEBUG") == "true"
	if err != nil {
		t.Fatalf("error opening stub database connection: %v", err)
	}
	defer db.Close()

	// mock rows
	rows := sqlmock.NewRows([]string{"id", "name", "email"}).
		AddRow(1, "John", "john@example.com").
		AddRow(2, "Doe", "doe@example.com")

	// Expect the query
	mock.ExpectQuery("SELECT id, name, email FROM users").WillReturnRows(rows)

	// Create a request
	req, err := http.NewRequest("GET", "/api/go/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(getUsers(db, debugMode))
	handler.ServeHTTP(rr, req)

	// Check if the status code is OK
	assert.Equal(t, http.StatusOK, rr.Code)

	// Decode the response body
	var users []User
	err = json.NewDecoder(rr.Body).Decode(&users)
	assert.Nil(t, err)

	// Assert that we received two users
	assert.Equal(t, 2, len(users))
	assert.Equal(t, "John", users[0].Name)
	assert.Equal(t, "doe@example.com", users[1].Email)
}

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	debugMode := os.Getenv("DEBUG") == "true"
	if err != nil {
		t.Fatalf("error opening stub database connection: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("Alice", "alice@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	user := User{Name: "Alice", Email: "alice@example.com"}
	userBytes, _ := json.Marshal(user)

	req, err := http.NewRequest("POST", "/api/go/users", bytes.NewBuffer(userBytes))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(createUser(db, debugMode))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var createdUser User
	err = json.NewDecoder(rr.Body).Decode(&createdUser)
	assert.Nil(t, err)

	assert.Equal(t, 1, createdUser.Id)
	assert.Equal(t, "Alice", createdUser.Name)
	assert.Equal(t, "alice@example.com", createdUser.Email)
}

func TestDeleteUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	debugMode := os.Getenv("DEBUG") == "true"
	if err != nil {
		t.Fatalf("error opening stub database connection: %v", err)
	}
	defer db.Close()

	// Expect a query to get the user, then to delete the user
	mock.ExpectQuery("SELECT id, name, email FROM users WHERE id = \\$1").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).AddRow(1, "John", "john@example.com"))
	mock.ExpectExec("DELETE FROM users WHERE id = \\$1").WithArgs("1").WillReturnResult(sqlmock.NewResult(1, 1))

	req, err := http.NewRequest("DELETE", "/api/go/users/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/api/go/users/{id}", deleteUser(db, debugMode)).Methods("DELETE")
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `"User deleted successfully"`, strings.TrimSpace(rr.Body.String()))
}

func TestEnableCORS(t *testing.T) {
	// Prepare a request with OPTIONS method for CORS
	req, err := http.NewRequest("OPTIONS", "/api/go/users", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := enableCORS(http.NotFoundHandler())

	// Serve the request
	handler.ServeHTTP(rr, req)

	// Check the CORS headers
	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}
