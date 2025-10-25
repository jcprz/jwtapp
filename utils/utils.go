package utils

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jcprz/jwtapp/models"
	"golang.org/x/crypto/bcrypt"
)

func ResponseJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)

}

func RespondWithError(w http.ResponseWriter, status int, message string) {
	var error models.Error
	error.Message = message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(error)

}

func GenerateToken(user models.User) (string, error) {
	secret := os.Getenv("SECRET")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
		"iss":   "course",
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		"iat":   time.Now().Unix(),
	})

	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil

}

func ComparePasswords(hashedPassword string, password []byte) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))

	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
