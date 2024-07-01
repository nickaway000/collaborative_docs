package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// Generate a JWT token for a given userID
func generateJWT(userID int) (string, error) {
    claims := &jwt.RegisteredClaims{
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
        Issuer:    fmt.Sprintf("%d", userID),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        return "", err
    }

    return tokenString, nil
}


// Hash password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Compare hashed password with plain text
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Parse form error", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE email=$1)", email).Scan(&exists)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if exists {
		errorMessage := "Email already exists"
		http.Redirect(w, r, "/register.html?error="+errorMessage, http.StatusSeeOther)
		return
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	var id int
	err = db.QueryRow("INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id", email, hashedPassword).Scan(&id)
	if err != nil {
		log.Printf("Error inserting user into database: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	token, err := generateJWT(id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("Static/login.html")
		if err != nil {
			http.Error(w, "Error loading login page", http.StatusInternalServerError)
			log.Printf("Error loading login page: %v\n", err)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	var dbPassword string
	var userID int
	err := db.QueryRow("SELECT id, password FROM users WHERE email = $1", email).Scan(&userID, &dbPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No user found with email: %s\n", email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			log.Printf("Error querying database: %v\n", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

	if !checkPasswordHash(password, dbPassword) {
		log.Printf("Password mismatch for user: %s\n", email)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:  "token",
		Value: token,
		Path:  "/",
	})

	documentServiceURL := fmt.Sprintf("http://localhost:8081/index.html?token=%s", token)
    http.Redirect(w, r, documentServiceURL, http.StatusSeeOther)
}
