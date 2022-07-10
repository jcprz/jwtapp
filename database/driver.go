package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

var db *sql.DB

func ConnectDB() *sql.DB {
	pgHost := os.Getenv("DB_HOST")
	pgUser := os.Getenv("DB_USER")
	pgPass := os.Getenv("DB_PASSWORD")
	pgPort := os.Getenv("DB_PORT")
	pgDbName := os.Getenv("DB_NAME")

	pgToStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", pgHost, pgPort, pgUser, pgPass, pgDbName)

	log.Printf("DB connection details: %s", pgToStr)

	db, _ = sql.Open("postgres", pgToStr)

	er := db.Ping()
	if er != nil {
		log.Fatal(er)
	}
	log.Println("Successfully connected to the Database")

	return db
}

func EnsureTableExists(db *sql.DB) error {
	log.Println("Executing DML for table creation/verification")
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS USERS (ID  SERIAL PRIMARY KEY, EMAIL VARCHAR(50), PASSWORD VARCHAR(100));")

	if err != nil {
		log.Panicf("Cannot create table. Error: %s", err)
	}
	log.Println("Table is created")
	return nil
}
