package main

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/subosito/gotenv"
)

func init() {
	gotenv.Load()
}

func main() {
	port := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	a := App{}
	a.Initialize()

	a.Run(port)
}
