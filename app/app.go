package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/andrewbenington/go-spotify/config"
	"github.com/andrewbenington/go-spotify/controller"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

type App struct {
	Router     *mux.Router
	Controller *controller.Controller
}

func (a *App) Initialize() {
	a.Controller = &controller.Controller{}
	a.initRouter()
}

func (a *App) initDB() {
	cfg := config.GetConfig()
	conn, err := pgx.Connect(context.Background(), cfg.GetDBString())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	defer conn.Close(context.Background())
}

func (a *App) initRouter() {
	a.Router = mux.NewRouter()

	a.Router.HandleFunc("/room", a.Controller.GetAllRooms).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room", a.Controller.CreateRoom).Methods("POST")

	a.Router.HandleFunc("/{code}", a.Controller.GetRoom).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/{code}/queue", a.Controller.GetQueue).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/{code}/queue/{song}", a.Controller.PushToQueue).Methods("POST")
	a.Router.HandleFunc("/{code}/search", a.Controller.Search).Methods("GET", "OPTIONS")
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, corsMW(logMW(a.Router))))
}

func logMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s - %s (%s)", r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(w, r)
	})
}

func corsMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, *")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
