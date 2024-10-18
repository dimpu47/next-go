package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// User struct
type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// DB connection
func main() {
	// Enable debug mode based on environment variable
	debugMode := os.Getenv("DEBUG") == "true"
	if debugMode {
		log.Println("Debug mode enabled")
	}

	// Connect to the database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create a table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT)")
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	// create router with mux
	router := mux.NewRouter()
	router.HandleFunc("/api/go/users", getUsers(db, debugMode)).Methods("GET")
	router.HandleFunc("/api/go/users/{id}", getUser(db, debugMode)).Methods("GET")
	router.HandleFunc("/api/go/users", createUser(db, debugMode)).Methods("POST")
	router.HandleFunc("/api/go/users/{id}", updateUser(db, debugMode)).Methods("PUT")
	router.HandleFunc("/api/go/users/{id}", deleteUser(db, debugMode)).Methods("DELETE")

	// Wrap the router with CORS and JSON content type middleware
	enhancedRouter := enableCORS(jsonContentTypeMiddleware(router))

	// Start the server
	log.Printf("Server started on http://0.0.0.0:8000 (Debug mode: %v)", debugMode)
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", enhancedRouter))
}

// enableCORS middleware
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// jsonContentTypeMiddleware middleware
func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// getUsers handler with logging
func getUsers(db *sql.DB, debug bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if debug {
			log.Println("Fetching all users")
		}
		rows, err := db.Query("SELECT id, name, email FROM users")
		if err != nil {
			log.Printf("Error fetching users: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.Id, &user.Name, &user.Email); err != nil {
				log.Printf("Error scanning row: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			users = append(users, user)
		}
		json.NewEncoder(w).Encode(users)
		if debug {
			log.Println("Successfully fetched users")
		}
	}
}

// getUser handler with logging
func getUser(db *sql.DB, debug bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		if debug {
			log.Printf("Fetching user with ID: %s", id)
		}

		var user User
		err := db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id).Scan(&user.Id, &user.Name, &user.Email)
		if err != nil {
			log.Printf("Error fetching user with ID %s: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(user)
		if debug {
			log.Printf("Successfully fetched user with ID: %s", id)
		}
	}
}

// createUser handler with logging
func createUser(db *sql.DB, debug bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if debug {
			log.Println("Creating new user")
		}

		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			log.Printf("Error decoding request body: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = db.QueryRow("INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id", user.Name, user.Email).Scan(&user.Id)
		if err != nil {
			log.Printf("Error inserting new user: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(user)
		if debug {
			log.Printf("Successfully created user with ID: %d", user.Id)
		}
	}
}

// updateUser handler with logging
func updateUser(db *sql.DB, debug bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		if debug {
			log.Printf("Updating user with ID: %s", id)
		}

		var user User
		json.NewDecoder(r.Body).Decode(&user)

		_, err := db.Exec("UPDATE users SET name = $1, email = $2 WHERE id = $3", user.Name, user.Email, id)
		if err != nil {
			log.Printf("Error updating user with ID %s: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var updatedUser User
		err = db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id).Scan(&updatedUser.Id, &updatedUser.Name, &updatedUser.Email)
		if err != nil {
			log.Printf("Error fetching updated user with ID %s: %v", id, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(updatedUser)
		if debug {
			log.Printf("Successfully updated user with ID: %s", id)
		}
	}
}

// deleteUser handler with logging
func deleteUser(db *sql.DB, debug bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]
		if debug {
			log.Printf("Deleting user with ID: %s", id)
		}

		var user User
		err := db.QueryRow("SELECT id, name, email FROM users WHERE id = $1", id).Scan(&user.Id, &user.Name, &user.Email)
		if err != nil {
			log.Printf("Error fetching user with ID %s: %v", id, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, err = db.Exec("DELETE FROM users WHERE id = $1", id)
		if err != nil {
			log.Printf("Error deleting user with ID %s: %v", id, err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode("User deleted successfully")
		if debug {
			log.Printf("Successfully deleted user with ID: %s", id)
		}
	}
}
