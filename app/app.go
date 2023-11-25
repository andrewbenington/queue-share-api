package app

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/controller"
	"github.com/andrewbenington/queue-share-api/user"
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
	a.Router.HandleFunc("/room/{code}", a.Controller.GetRoom).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}", a.Controller.DeleteRoom).Methods("DELETE", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/queue", a.Controller.GetQueue).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/queue/{song}", a.Controller.PushToQueue).Methods("POST", "OPTIONS")

	a.Router.HandleFunc("/room/{code}/play", a.Controller.Play).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/pause", a.Controller.Pause).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/next", a.Controller.Next).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/previous", a.Controller.Previous).Methods("POST", "OPTIONS")

	a.Router.HandleFunc("/room/{code}/devices", a.Controller.Devices).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/playlists", a.Controller.UserPlaylists).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/room/{code}/search", a.Controller.Search).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/guest", a.Controller.AddGuest).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/guests_and_members", a.Controller.GetRoomGuestsAndMembers).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/join", a.Controller.JoinRoomAsMember).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/permissions", a.Controller.GetRoomPermissions).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/moderator", a.Controller.SetModerator).Methods("PUT", "OPTIONS")

	a.Router.HandleFunc("/user", a.Controller.CreateUser).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/user", a.Controller.CurrentUser).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/user/room", a.Controller.GetUserRoom).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/auth/token", a.Controller.GetToken).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/auth/spotify-url", a.Controller.GetSpotifyLoginURL).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/auth/spotify-redirect", a.Controller.SpotifyAuthRedirect).Methods("GET", "OPTIONS")
}

func (a *App) Run(addr string) {
	log.Printf("serving on %s...", addr)
	log.Fatalf("server error: %s", http.ListenAndServe(addr, corsMW(authMW(logMW(a.Router)))))
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

type ErrorResponse struct {
	Error string `json:"error"`
}

func authMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) < 2 {
			next.ServeHTTP(w, r)
			return
		}
		reqToken = splitToken[1]

		id, err := user.GetTokenID(reqToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{err.Error()})
			return
		}
		ctx := context.WithValue(r.Context(), auth.UserContextKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
