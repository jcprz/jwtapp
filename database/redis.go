package database

import (
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis"
	// Redis 6
)

var rds *redis.Client

func ConnectRedis() *redis.Client {

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	rdsToStr := fmt.Sprintf("host=%s port=%s password=%s sslmode=disable\n", redisHost, redisPort, redisPassword)

	log.Printf("Redis connection details: %s", rdsToStr)
	redisAddr := fmt.Sprintf("%s:%s", redisHost, redisPort)

	client := redis.NewClient(&redis.Options{

		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0, //  default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("Unable to acquire connection with Redis: %s", err)
	}
	log.Println("Successflly connected to Redis")

	return client
}
