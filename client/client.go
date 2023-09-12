package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/andrewbenington/go-spotify/auth"
	"github.com/andrewbenington/go-spotify/db"
	"github.com/andrewbenington/go-spotify/room"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

func FromRequest(r *http.Request) (statusCode int, client *spotify.Client, err error) {
	ctx := r.Context()
	code, password := room.ParametersFromRequest(r)
	if code == "" {
		return http.StatusBadRequest, nil, fmt.Errorf("invalid room code")
	}
	status, token, err := DecryptRoomToken(ctx, code, password)
	if err != nil {
		return status, nil, err
	}

	authenticator := spotifyauth.New(spotifyauth.WithScopes(auth.Scopes...))
	httpClient := authenticator.Client(ctx, token)
	spotifyClient := spotify.New(httpClient)

	// refresh token if stale
	if token.Expiry.Before(time.Now()) {
		_, err := spotifyClient.CurrentUser(ctx)
		if err != nil {
			return http.StatusInternalServerError, nil, fmt.Errorf("refresh token request: %w", err)
		}

		transport, ok := httpClient.Transport.(*oauth2.Transport)
		if !ok {
			return http.StatusInternalServerError, nil, errors.New("get token source")
		}
		newToken, err := transport.Source.Token()
		if err != nil {
			return http.StatusInternalServerError, nil, fmt.Errorf("get refreshed token: %w", err)
		}
		encryptedAccessToken, err := auth.EncryptToken(newToken.AccessToken, password)
		if err != nil {
			return http.StatusUnauthorized, nil, fmt.Errorf("invalid password")
		}
		encryptedRefreshToken, err := auth.EncryptToken(newToken.RefreshToken, password)
		if err != nil {
			return http.StatusUnauthorized, nil, fmt.Errorf("invalid password")
		}
		err = db.Service().RoomStore.UpdateEncryptedRoomTokens(ctx, code, encryptedAccessToken, newToken.Expiry, encryptedRefreshToken)
		if err != nil {
			return http.StatusInternalServerError, nil, fmt.Errorf("update room tokens: %w", err)
		}
	}

	return http.StatusOK, spotifyClient, nil
}

func DecryptRoomToken(ctx context.Context, code string, password string) (int, *oauth2.Token, error) {
	encrytpedAccessToken, accessTokenExpiry, encryptedRefreshToken, err := db.Service().RoomStore.GetEncryptedRoomTokens(ctx, code)
	if err == sql.ErrNoRows {
		return http.StatusNotFound, nil, fmt.Errorf("No room with code '%s'", code)
	}
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
