package app

import (
	"log"
	"net/http"

	"github.com/andrewbenington/go-spotify/auth"
	"github.com/andrewbenington/go-spotify/controller"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type App struct {
	Router     *mux.Router
	Controller *controller.Controller
}

func (a *App) Initialize() {
	session := uuid.New()
	client := auth.AuthenticateUser(session.String(), false)
	if client == nil {
		panic("error getting authenticated client")
	}
	a.Controller = controller.NewController(client)
	a.initRouter()
}

func (a *App) initRouter() {
	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/queue", a.Controller.GetQueue).Methods("GET")
	a.Router.HandleFunc("/queue/{song}", a.Controller.PushToQueue).Methods("POST")
	a.Router.HandleFunc("/search", a.Controller.Search).Methods("GET")
	a.Router.HandleFunc("/auth", a.Controller.Auth).Methods("GET")
	a.Router.HandleFunc("/auth/redirect", a.Controller.RedirectHandler).Methods("GET")
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
		next.ServeHTTP(w, r)
	})
}
