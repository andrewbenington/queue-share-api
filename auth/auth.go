package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var (
	auth    *spotifyauth.Authenticator
	session string
	c       chan *spotify.Client
	Scopes  = []string{
		spotifyauth.ScopeUserReadPrivate,
		spotifyauth.ScopePlaylistModifyPrivate,
		spotifyauth.ScopePlaylistModifyPublic,
		spotifyauth.ScopeUserReadPlaybackState,
		spotifyauth.ScopeUserModifyPlaybackState,
	}
	client *spotify.Client
	user   *spotify.PrivateUser
)

func Client() *spotify.Client {
	return client
}

func AuthenticateUser(s string, refresh bool) *spotify.Client {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Couldn't get environment: %s\n", err)
	}
	if !refresh && os.Getenv("ACCESS_TOKEN") != "" {
		fmt.Println("Authenticating with existing token...")
		client, err = authenticateWithToken()
		if err == nil {
			return client
		}

		fmt.Printf("Error with existing token: %s\n", err)
	}
	if os.Getenv("SPOTIFY_ID") == "" {
		fmt.Println("Environment missing SPOTIFY_ID")
		os.Exit(1)
	}
	if os.Getenv("SPOTIFY_SECRET") == "" {
		fmt.Println("Environment missing SPOTIFY_SECRET")
		os.Exit(1)
	}
	session = s
	fmt.Printf("Authenticating session %s\n", s)
	c = make(chan *spotify.Client)
	auth = spotifyauth.New(spotifyauth.WithRedirectURL("http://localhost:5757"), spotifyauth.WithScopes(Scopes...))

	redirectServer := startRedirectServer()

	url := auth.AuthURL(session)
	fmt.Printf("authenticate with url:\n%s\n", url)
	exec.Command("open", url).Run()
	client = <-c
	redirectServer.Shutdown(context.Background())
	return client
}

func authenticateWithToken() (*spotify.Client, error) {
	expiry, err := time.Parse(time.RFC3339, os.Getenv("TOKEN_EXPIRY"))
	if err != nil {
		return nil, err
	}
	token := &oauth2.Token{
		AccessToken:  os.Getenv("ACCESS_TOKEN"),
		RefreshToken: os.Getenv("REFRESH_TOKEN"),
		Expiry:       expiry,
	}
	auth = spotifyauth.New(spotifyauth.WithScopes(Scopes...))
	httpClient := auth.Client(context.Background(), token)
	spotifyClient := spotify.New(httpClient)

	_, err = spotifyClient.CurrentUser(context.Background())
	if err != nil {
		return nil, err
	}

	transport, ok := httpClient.Transport.(*oauth2.Transport)
	if !ok {
		return nil, errors.New("Error getting token source")
	}
	currentToken, err := transport.Source.Token()
	if err != nil {
		return nil, err
	}
	updateDotEnvToken(currentToken)
	return spotifyClient, err
}

func startRedirectServer() *http.Server {
	srv := &http.Server{Addr: ":5757"}
	http.HandleFunc("/", RedirectHandler)
	go func() {
		err := srv.ListenAndServe()
		if err == http.ErrServerClosed {
			fmt.Println("Server closed successfully")
			return
		}
		if err != nil {
			fmt.Println("Error with server:", err)
		}
	}()
	return srv
}

// the user will eventually be redirected back to your redirect URL
// typically you'll have a handler set up like the following:
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	// use the same state string here that you used to generate the URL
	token, err := auth.Token(r.Context(), session, r)
	if err != nil {
		fmt.Printf("Couldn't get token: %s\n", err)
		return
	}
	err = updateDotEnvToken(token)
	w.WriteHeader(200)
	w.Write([]byte("Successful. You may close this tab"))
	// create a client using the specified token
	client := spotify.New(auth.Client(r.Context(), token))
	c <- client
	// the client can now be used to make authenticated requests
}

func updateDotEnvToken(token *oauth2.Token) error {
	env, err := godotenv.Read(".env")
	if err != nil {
		return fmt.Errorf("Error loading .env file: %w", err)
	}
	env["ACCESS_TOKEN"] = token.AccessToken
	env["REFRESH_TOKEN"] = token.RefreshToken
	env["TOKEN_EXPIRY"] = token.Expiry.Format(time.RFC3339)
	godotenv.Write(env, ".env")

	fmt.Println("Tokens added to .env successfully.")
	return nil
}
