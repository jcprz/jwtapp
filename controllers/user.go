package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"jwt-app/models"
	userRepository "jwt-app/repository/user"
	"jwt-app/utils"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"golang.org/x/crypto/bcrypt"
)

func (c Controller) Signup(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user models.User

		json.NewDecoder(r.Body).Decode(&user)
		log.Println(user)

		if user.Email == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email is missing.")
			return
		}

		if user.Password == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Password is missing.")
			return

		}

		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)

		if err != nil {
			log.Fatal(err)
		}

		user.Password = string(hash)

		userRepo := userRepository.UserRepository{}
		user = userRepo.Signup(db, user)

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Server Error.")
			return
		}

		user.Password = ""
		utils.ResponseJSON(w, http.StatusCreated, user)
	}

}

func (c Controller) Login(db *sql.DB, redis *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var user models.User
		var jwt models.JWT

		json.NewDecoder(r.Body).Decode(&user)

		if user.Email == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email is missing.")
		}

		if user.Password == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Password is missing.")
		}

		password := user.Password

		userRepo := userRepository.UserRepository{}

		user, err := userRepo.Login(db, redis, user)

		hashedPassword := user.Password

		token, err := utils.GenerateToken(user)

		if err != nil {
			log.Fatal(err)
		}

		isValidPasswd := utils.ComparePasswords(hashedPassword, []byte(password))

		if isValidPasswd {
			w.Header().Set("Authorization", token)

			jwt.Token = token
			utils.ResponseJSON(w, http.StatusOK, jwt)

		} else {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid token.")
		}

	}

}

func (c Controller) Delete(db *sql.DB, redis *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var user models.User

		json.NewDecoder(r.Body).Decode(&user)

		if user.Email == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email is missing.")
		}

		userRepo := userRepository.UserRepository{}

		err := userRepo.Delete(db, redis, user)

		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			// log.Fatal(err)
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
		} else {

			utils.ResponseJSON(w, http.StatusOK, "User has been deleted")
		}

	}

}

func (c Controller) TokenVerifyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearerToken := r.Header.Get("Authorization")
		var authHeader string

		if bearerToken != "" {
			authHeader = strings.Split(bearerToken, " ")[1]
		}

		if authHeader != "" {

		}

		if len(authHeader) > 2 {
			token, error := jwt.Parse(authHeader, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}

				return []byte(os.Getenv("SECRET")), nil
			})

			if error != nil {
				utils.RespondWithError(w, http.StatusUnauthorized, error.Error())
				return
			}

			if token.Valid {
				next.ServeHTTP(w, r)
			} else {
				utils.RespondWithError(w, http.StatusUnauthorized, error.Error())
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid Token")
			return
		}
	})
}
