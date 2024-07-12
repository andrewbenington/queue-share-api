package app

import (
	"log"
	"net/http"

	"github.com/andrewbenington/queue-share-api/controller"
	"github.com/gorilla/mux"
)

type App struct {
	Router          *mux.Router
	Controller      *controller.Controller
	StatsController *controller.StatsController
}

func (a *App) Initialize() {
	a.Controller = &controller.Controller{}
	a.initRouter()
}

func (a *App) Run(addr string) {
	log.Printf("serving on %s...", addr)
	log.Fatalf("server error: %s", http.ListenAndServe(addr, withMiddleware(a.Router)))
}
