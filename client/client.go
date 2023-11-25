package client

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

func ForRoom(ctx context.Context, code string) (statusCode int, client *spotify.Client, err error) {
	encrytpedAccessToken, accessTokenExpiry, encryptedRefreshToken, err := db.Service().RoomStore.GetEncryptedRoomTokens(ctx, code)
	if err == sql.ErrNoRows {
		return http.StatusNotFound, nil, fmt.Errorf("Room credentials not found")
	}

	status, token, err := DecryptRoomToken(ctx, encrytpedAccessToken, accessTokenExpiry, encryptedRefreshToken)
	if err != nil {
		return status, nil, err
	}
	log.Printf(token.AccessToken)
	if err != nil {
		log.Printf("Error decrypting room token: %s", err)
		return http.StatusInternalServerError, nil, err
	}

	authenticator := spotifyauth.New(spotifyauth.WithScopes(auth.SpotifyScopes...))
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
		err = db.Service().RoomStore.UpdateSpotifyToken(ctx, code, newToken)
		if err != nil {
			return http.StatusInternalServerError, nil, fmt.Errorf("update room tokens: %w", err)
		}
	}

	return http.StatusOK, spotifyClient, nil
}

func DecryptRoomToken(ctx context.Context, encrytpedAccessToken []byte, accessTokenExpiry time.Time, encryptedRefreshToken []byte) (int, *oauth2.Token, error) {
	decryptedAccessToken, err := auth.AESGCMDecrypt(encrytpedAccessToken)
	if err != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf("decryption error")
	}
	decryptedRefreshToken, err := auth.AESGCMDecrypt(encryptedRefreshToken)
	if err != nil {
		return http.StatusInternalServerError, nil, fmt.Errorf("decryption error")
	}
	token := &oauth2.Token{
		AccessToken:  decryptedAccessToken,
		Expiry:       accessTokenExpiry,
		RefreshToken: decryptedRefreshToken,
	}
	return http.StatusOK, token, nil
}
