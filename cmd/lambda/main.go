package main

import (
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gorillamux"
	_ "github.com/lib/pq"
	"github.com/subosito/gotenv"

	"github.com/jcprz/jwtapp/pkg/app"
)

var muxAdapter *gorillamux.GorillaMuxAdapter

func init() {
	// Load .env file if present (for local testing)
	gotenv.Load()

	// Initialize the app
	application := app.App{}
	application.Initialize()

	// Create the Lambda adapter for gorilla/mux
	muxAdapter = gorillamux.New(application.Router)
}

func main() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// Running in Lambda - use the adapter's Proxy method directly
		lambda.Start(muxAdapter.Proxy)
	} else {
		// Running locally for testing
		log.Println("Lambda handler initialized. Use SAM or Lambda emulator to test.")
	}
}
