package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/andrewbenington/go-spotify/auth"
	"github.com/andrewbenington/go-spotify/db"
	"github.com/andrewbenington/go-spotify/room"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

func FromRequest(r *http.Request) (statusCode int, client *spotify.Client, err error) {
	code, password := room.ParametersFromRequest(r)
	if code == "" {
		return http.StatusBadRequest, nil, fmt.Errorf("invalid room code")
	}
	status, token, err := DecryptRoomToken(r.Context(), code, password)
	if err != nil {
		return status, nil, err
	}

	authenticator := spotifyauth.New(spotifyauth.WithScopes(auth.Scopes...))
	httpClient := authenticator.Client(r.Context(), token)
	fmt.Println(httpClient)
	spotifyClient := spotify.New(httpClient)

	if err != nil {
		return http.StatusUnauthorized, nil, err
	}
	return http.StatusOK, spotifyClient, nil
}

func DecryptRoomToken(ctx context.Context, code string, password string) (int, *oauth2.Token, error) {
	encrytpedAccessToken, accessTokenExpiry, encryptedRefreshToken, err := db.Service().RoomStore.GetEncryptedRoomTokens(ctx, code)
	if err != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf("get encrypted room token: %w", err)
	}
	decryptedAccessToken, err := auth.DecryptToken(encrytpedAccessToken, password)
	if err != nil {
		return http.StatusUnauthorized, nil, fmt.Errorf("invalid password")
	}
	decryptedRefreshToken, err := auth.DecryptToken(encryptedRefreshToken, password)
	if err != nil {
		return http.StatusUnauthorized, nil, fmt.Errorf("invalid password")
	}
	token := &oauth2.Token{
		AccessToken:  decryptedAccessToken,
		Expiry:       accessTokenExpiry,
		RefreshToken: decryptedRefreshToken,
	}
	return http.StatusOK, token, nil
}
