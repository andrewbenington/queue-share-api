package app

import (
	"log"
	"net/http"

	"github.com/andrewbenington/queue-share-api/controller"
	"github.com/gorilla/mux"
)

type App struct {
	Router     *mux.Router
	Controller *controller.Controller
}

func (a *App) Initialize() {
	a.Controller = &controller.Controller{}
	a.initRouter()
}

func (a *App) initRouter() {
	a.Router = mux.NewRouter()

	// health
	a.Router.HandleFunc("/health", a.Controller.Health).Methods("GET", "OPTIONS")

	// a.Router.HandleFunc("/room", a.Controller.GetAllRooms).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room", a.Controller.CreateRoom).Methods("POST", "OPTIONS")

	a.Router.HandleFunc("/{code}", a.Controller.GetRoom).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/{code}/queue", a.Controller.GetQueue).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/{code}/queue/{song}", a.Controller.PushToQueue).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/{code}/search", a.Controller.Search).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/user", a.Controller.CreateUser).Methods("POST", "OPTIONS")
}

func (a *App) Run(addr string) {
	log.Printf("serving on %s...", addr)
	log.Fatalf("server error: %s", http.ListenAndServe(addr, corsMW(logMW(a.Router))))
}

func logMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			log.Printf("%s - %s (%s)", r.Method, r.URL.Path, r.RemoteAddr)
		}

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
