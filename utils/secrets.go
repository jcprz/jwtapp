package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// GetSecretValue retrieves a secret value from AWS Secrets Manager
func GetSecretValue(secretArn string) (string, error) {
	if secretArn == "" {
		return "", fmt.Errorf("secret ARN is empty")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretArn),
	}

	result, err := client.GetSecretValue(context.TODO(), input)
	if err != nil {
		return "", fmt.Errorf("unable to retrieve secret: %w", err)
	}

	return *result.SecretString, nil
}

// GetDBPasswordFromSecret retrieves the database password from Secrets Manager
func GetDBPasswordFromSecret() string {
	secretArn := os.Getenv("DB_PASSWORD_SECRET_ARN")
	if secretArn == "" {
		// Fallback to environment variable if not using Secrets Manager
		log.Println("DB_PASSWORD_SECRET_ARN not set, using DB_PASSWORD from environment")
		return os.Getenv("DB_PASSWORD")
	}

	secretString, err := GetSecretValue(secretArn)
	if err != nil {
		log.Printf("Error retrieving DB password from Secrets Manager: %v", err)
		// Fallback to environment variable
		return os.Getenv("DB_PASSWORD")
	}

	// Parse the secret JSON to extract the password
	var secretData map[string]interface{}
	if err := json.Unmarshal([]byte(secretString), &secretData); err != nil {
		log.Printf("Error parsing DB secret JSON: %v", err)
		return os.Getenv("DB_PASSWORD")
	}

	if password, ok := secretData["password"].(string); ok {
		return password
	}

	log.Println("Password not found in secret, using DB_PASSWORD from environment")
	return os.Getenv("DB_PASSWORD")
}

// GetJWTSecretFromSecret retrieves the JWT secret from Secrets Manager
func GetJWTSecretFromSecret() string {
	secretArn := os.Getenv("JWT_SECRET_ARN")
	if secretArn == "" {
		// Fallback to environment variable if not using Secrets Manager
		log.Println("JWT_SECRET_ARN not set, using SECRET from environment")
		return os.Getenv("SECRET")
	}

	secretString, err := GetSecretValue(secretArn)
	if err != nil {
		log.Printf("Error retrieving JWT secret from Secrets Manager: %v", err)
		// Fallback to environment variable
		return os.Getenv("SECRET")
	}

	return secretString
}
