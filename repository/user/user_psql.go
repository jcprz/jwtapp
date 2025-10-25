package userRepository

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/jcprz/jwtapp/models"
)

type UserRepository struct{}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (u UserRepository) Signup(db *sql.DB, user models.User) models.User {
	err := db.QueryRow("insert into users (email, password) values ($1, $2) RETURNING id;", user.Email, user.Password).Scan(&user.ID)

	if err != nil {
		log.Printf("Error creating user: %v", err)
	}

	user.Password = ""

	return user
}

func (u UserRepository) Login(db *sql.DB, redis *redis.Client, user models.User) (models.User, error) {
	result, err := redis.HGetAll(user.Email).Result()

	if err != nil || len(result) == 0 {
		log.Printf("Unable to find %s on redis cache", user.Email)
		log.Println("Executing query to the database")
		row := db.QueryRow("select * from users where email = $1;", user.Email)
		err := row.Scan(&user.ID, &user.Email, &user.Password)

		if err != nil {
			return user, err
		}

		log.Printf("User %s found in the database\n", user.Email)

		log.Println("Caching user on Redis (without password)")
		// SECURITY FIX: Do NOT cache the password in Redis
		redis.HSet(user.Email, "id", user.ID)
		redis.HSet(user.Email, "email", user.Email)

		data, err := redis.HGetAll(user.Email).Result()

		log.Println(data)

		return user, nil

	}

	// User found in cache, need to get password from database
	log.Printf("Cache hit for email: %s. Fetching password from database.\n", user.Email)

	// Get ID from cache
	if idStr, ok := result["id"]; ok {
		if id, err := strconv.Atoi(idStr); err == nil {
			user.ID = id
		}
	}
	user.Email = result["email"]

	// Always fetch password from database (never cache it)
	row := db.QueryRow("select password from users where email = $1;", user.Email)
	err = row.Scan(&user.Password)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (u UserRepository) Delete(db *sql.DB, redis *redis.Client, user models.User) error {

	log.Printf("Deleting user: %s from the database", user.Email)
	delUser := db.QueryRow("delete from users where email = $1 RETURNING id;", user.Email)
	err := delUser.Scan(&user.ID)

	if err != nil {
		log.Printf("User %s not found on the database\n", user.Email)
		return err
	}

	log.Printf("User %s has been deleted from the database\n", user.Email)

	// Delete user from redis too
	redis.Del(user.Email)
	log.Println("User has been deleted from redis cache")
	return nil
}
