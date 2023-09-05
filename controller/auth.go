package controller

import (
	"fmt"
	"net/http"
	"os"

	"github.com/andrewbenington/go-spotify/auth"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	sauth *spotifyauth.Authenticator
)

// the user will eventually be redirected back to your redirect URL
// typically you'll have a handler set up like the following:
func (c *Controller) Auth(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Couldn't get environment: %s\n", err)
	}
	if os.Getenv("SPOTIFY_ID") == "" {
		fmt.Println("Environment missing SPOTIFY_ID")
		os.Exit(1)
	}
	if os.Getenv("SPOTIFY_SECRET") == "" {
		fmt.Println("Environment missing SPOTIFY_SECRET")
		os.Exit(1)
	}
	c.NewSession = uuid.New().String()
	fmt.Printf("Authenticating session %s\n", c.NewSession)
	sauth = spotifyauth.New(spotifyauth.WithRedirectURL("http://205.178.61.21:3001/auth/redirect"), spotifyauth.WithScopes(auth.Scopes...))

	url := sauth.AuthURL(c.NewSession)
	responseBody := fmt.Sprintf(`
<head>
  <meta http-equiv="Refresh" content="0; URL=%s" />
</head>
	`, url)
	w.Write([]byte(responseBody))
}

// the user will eventually be redirected back to your redirect URL
// typically you'll have a handler set up like the following:
func (c *Controller) RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// use the same state string here that you used to generate the URL
	token, err := sauth.Token(r.Context(), c.NewSession, r)
	if err != nil {
		fmt.Printf("Couldn't get token: %s\n", err)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte("Successful. You may close this tab"))
	// create a client using the specified token
	client := spotify.New(sauth.Client(r.Context(), token))
	c.Client = client
	// the client can now be used to make authenticated requests
}
