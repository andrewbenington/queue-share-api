package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/config"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/andrewbenington/queue-share-api/user"
	"github.com/google/uuid"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type TokenResponse struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	User      *user.User `json:"user"`
}

func (c *Controller) GetToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	username, password, ok := r.BasicAuth()
	if !ok {
		requests.RespondAuthError(w)
		return
	}

	authenticated, err := user.Authenticate(r.Context(), tx, username, password)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("authenticate: %s", err)
		requests.RespondInternalError(w)
		return
	}
	if !authenticated || err == sql.ErrNoRows {
		requests.RespondAuthError(w)
		return
	}

	u, err := user.GetByUsername(r.Context(), tx, username)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	token, expiry, err := u.GetJWT()
	if err != nil {
		log.Printf("generate jwt: %s", err)
		requests.RespondInternalError(w)
		return
	}

	resp := TokenResponse{
		Token:     token,
		ExpiresAt: expiry,
		User:      u,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

type GetSpotifyLoginURLResponse struct {
	URL string `json:"url"`
}

func (c *Controller) GetSpotifyLoginURL(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		requests.RespondAuthError(w)
		return
	}
	redirect := r.URL.Query().Get("redirect")

	state := uuid.New().String()
	auth.SpotifyStatesLock.Lock()
	auth.SpotifyStates[state] = auth.SpotifyLoginState{
		UserID:      userID,
		RedirectURI: redirect,
	}
	auth.SpotifyStatesLock.Unlock()

	authenticator := spotifyauth.New(spotifyauth.WithRedirectURL(config.GetSpotifyRedirect()), spotifyauth.WithScopes(auth.SpotifyScopes...))
	url := authenticator.AuthURL(state)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetSpotifyLoginURLResponse{URL: url})
}

func (c *Controller) SpotifyAuthRedirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	state := query.Get("state")
	if state == "" {
		log.Printf("no state in spotify request: %s", query.Get("error"))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no state in spotify request"))
		return
	}

	auth.SpotifyStatesLock.Lock()
	loginState, ok := auth.SpotifyStates[state]
	if ok {
		delete(auth.SpotifyStates, state)
	}
	auth.SpotifyStatesLock.Unlock()

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}

	if !ok {
		log.Printf("unknown spotify state")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unknown spotify state"))
		return
	}

	redirectURI, err := url.Parse(loginState.RedirectURI)
	if err != nil {
		log.Printf("bad redirect uri: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("bad redirect uri"))
		return
	}

	redirectQuery := redirectURI.Query()

	defer func() {
		redirectURI.RawQuery = redirectQuery.Encode()
		http.Redirect(w, r, redirectURI.String(), http.StatusSeeOther)
	}()

	spotifyError := query.Get("error")
	if spotifyError != "" {
		log.Printf("spotify error: %s", spotifyError)
		redirectQuery.Add("error", spotifyError)
		return
	}

	code := query.Get("code")
	if code == "" {
		log.Printf("no code in spotify request: %s", query.Get("error"))
		redirectQuery.Add("error", "No code present in Spotify request")
		return
	}

	authenticator := spotifyauth.New(spotifyauth.WithRedirectURL(config.GetSpotifyRedirect()), spotifyauth.WithScopes(auth.SpotifyScopes...))
	token, err := authenticator.Token(ctx, state, r)
	if err != nil {
		log.Printf("get spotify token: %s\n", err)
		redirectQuery.Add("error", fmt.Sprintf("Error getting Spotify token: %s", err))
		return
	}

	err = user.UpdateSpotifyToken(r.Context(), tx, loginState.UserID, token)
	if err != nil {
		log.Printf("update user spotify token: %s\n (loginState %+v)", err, loginState)
		redirectQuery.Add("error", fmt.Sprintf("Error updating user Spotify token: %s", err))
		return
	}

	spotifyClient := spotify.New(authenticator.Client(ctx, token))
	userData, err := service.GetUser(ctx, spotifyClient)
	if err != nil {
		log.Printf("get spotify user: %s\n", err)
		redirectQuery.Add("error", fmt.Sprintf("Error getting Spotify user: %s", err))
		return
	}

	err = user.UpdateSpotifyInfo(ctx, tx, loginState.UserID, userData)
	if err != nil {
		log.Printf("update user spotify info: %s\n", err)
		redirectQuery.Add("error", fmt.Sprintf("Error updating user Spotify info: %s", err))
		return
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("commit tx: %s\n", err)
		redirectQuery.Add("error", fmt.Sprintf("Error updating Spotify user info: %s", err))
		return
	}

	redirectQuery.Add("spotify_id", userData.ID)
	redirectQuery.Add("spotify_name", userData.Display)
	redirectQuery.Add("spotify_image", userData.ImageURL)
}
