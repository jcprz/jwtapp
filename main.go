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
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	a := App{}
	a.Initialize()

	a.Run(port)
}
