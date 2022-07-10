package main

import (
	"database/sql"
	"jwt-app/controllers"
	"jwt-app/database"
	"log"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

type App struct {
	DB     *sql.DB
	RDS    *redis.Client
	Router *mux.Router
}

func (a *App) Initialize() {
	a.DB = database.ConnectDB()
	a.RDS = database.ConnectRedis()
	a.Router = mux.NewRouter()
	database.EnsureTableExists(a.DB)

	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	controller := controllers.Controller{}
	a.Router.HandleFunc("/signup", controller.Signup(a.DB)).Methods("POST")
	a.Router.HandleFunc("/login", controller.Login(a.DB, a.RDS)).Methods("POST")
	a.Router.HandleFunc("/delete", controller.Delete(a.DB, a.RDS)).Methods("DELETE")
	a.Router.HandleFunc("/protected", controller.TokenVerifyMiddleware(controller.ProtectedEndpoint())).Methods("GET")
	a.Router.HandleFunc("/healthz", controller.HealthZ()).Methods("GET")
}

func (a *App) Run(port string) {
	log.Printf("Listening on port: %s\n", port[1:])
	log.Fatal(http.ListenAndServe(port, a.Router))
}
