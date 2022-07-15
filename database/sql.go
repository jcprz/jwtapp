package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

var db *sql.DB

func ConnectDB() *sql.DB {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbDialect := os.Getenv("DB_DIALECT")

	dbToStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)

	log.Printf("DB connection details: %s", dbToStr)

	db, _ = sql.Open(dbDialect, dbToStr)

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
