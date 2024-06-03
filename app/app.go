package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/controller"
	"github.com/andrewbenington/queue-share-api/user"
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
	a.Router.HandleFunc("/room/{code}/volume", a.Controller.SetVolume).Methods("PUT", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/player", a.Controller.GetPlayback).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/room/{code}/devices", a.Controller.Devices).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/playlists", a.Controller.UserPlaylists).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/suggested", a.Controller.SuggestedTracks).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/room/{code}/playlist", a.Controller.GetPlaylist).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/album", a.Controller.GetAlbum).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/artist", a.Controller.GetArtist).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/room/{code}/search", a.Controller.SearchTracksFromRoom).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/guest", a.Controller.AddGuest).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/guests-and-members", a.Controller.GetRoomGuestsAndMembers).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/join", a.Controller.JoinRoomAsMember).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/member", a.Controller.AddMember).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/member", a.Controller.DeleteMember).Methods("DELETE")
	a.Router.HandleFunc("/room/{code}/permissions", a.Controller.GetRoomPermissions).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/moderator", a.Controller.SetModerator).Methods("PUT", "OPTIONS")
	a.Router.HandleFunc("/room/{code}/password", a.Controller.UpdatePassword).Methods("PUT", "OPTIONS")

	a.Router.HandleFunc("/user", a.Controller.CreateUser).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/user", a.Controller.CurrentUser).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/user/rooms/hosted", a.Controller.GetUserHostedRooms).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/user/rooms/joined", a.Controller.GetUserJoinedRooms).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/user/spotify", a.Controller.UnlinkSpotify).Methods("DELETE", "OPTIONS")
	a.Router.HandleFunc("/user/has-spotify-history", a.Controller.UserHasSpotifyHistory).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/user/friend-suggestions", a.Controller.UserFriendRequestData).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/user/friend-request", a.Controller.UserSendFriendRequest).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/user/friend-request", a.Controller.UserDeleteFriendRequest).Methods("DELETE", "OPTIONS")
	a.Router.HandleFunc("/user/friends", a.Controller.UserGetFriends).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/user/friends", a.Controller.UserAcceptFriendRequest).Methods("POST", "OPTIONS")

	a.Router.HandleFunc("/stats/upload", a.StatsController.UploadHistory).Methods("POST", "OPTIONS")
	a.Router.HandleFunc("/stats/history", a.StatsController.GetAllHistory).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/stats/tracks-by-year", a.StatsController.GetTopTracksByYear).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/stats/albums-by-year", a.StatsController.GetTopAlbumsByYear).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/stats/artists-by-year", a.StatsController.GetTopArtistsByYear).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/stats/all-track-streams", a.StatsController.GetAllStreamsByURI).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/stats/streams", a.StatsController.GetAllStreamsByURI).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/stats/track", a.StatsController.GetTrackStatsByURI).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/stats/artist", a.StatsController.GetArtistStatsByURI).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/stats/album", a.StatsController.GetAlbumStatsByURI).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/stats/compare-tracks", a.StatsController.UserCompareFriendTopTracks).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/stats/compare-artists", a.StatsController.UserCompareFriendTopArtists).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/rankings/track/{spotify_uri}", a.StatsController.GetTrackRankingsByURI).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/rankings/album/{spotify_uri}", a.StatsController.GetAlbumRankingsByURI).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/rankings/artist/{spotify_uri}", a.StatsController.GetArtistRankingsByURI).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/rankings/track", a.StatsController.GetTopTracksByTimeframe).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/rankings/artist", a.StatsController.GetTopArtistsByTimeframe).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/rankings/album", a.StatsController.GetTopAlbumsByTimeframe).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/spotify/search-tracks", a.Controller.SearchTracksByUser).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/spotify/search-artists", a.Controller.SearchArtistsByUser).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/spotify/artists-by-uri", a.Controller.GetArtistsByURIs).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/auth/token", a.Controller.GetToken).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/auth/spotify-url", a.Controller.GetSpotifyLoginURL).Methods("GET", "OPTIONS")
	a.Router.HandleFunc("/auth/spotify-redirect", a.Controller.SpotifyAuthRedirect).Methods("GET", "OPTIONS")

	a.Router.HandleFunc("/version", a.Controller.GetVersion).Methods("GET", "OPTIONS")
}

func (a *App) Run(addr string) {
	log.Printf("serving on %s...", addr)
	log.Fatalf("server error: %s", http.ListenAndServe(addr, timeoutMW(corsMW(authMW(logMW(a.Router))))))
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
		fmt.Printf("User %s\n", id)
		ctx := context.WithValue(r.Context(), auth.UserContextKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func timeoutMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
