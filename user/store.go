package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/db/gen"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) InsertUser(ctx context.Context, tx *sql.Tx, username string, displayName string, password string) (newUserID string, err error) {
	userUUID, err := gen.New(tx).UserInsertWithPassword(ctx, gen.UserInsertWithPasswordParams{
		Username:    username,
		DisplayName: displayName,
		UserPass:    password,
	})
	if err != nil {
		return "", err
	}
	return userUUID.String(), nil
}

func (s *Store) UpdateSpotifyToken(ctx context.Context, userID string, oauthToken *oauth2.Token) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("parse user UUID: %w", err)
	}
	encryptedAccessToken, err := auth.AESGCMEncrypt(oauthToken.AccessToken)
	if err != nil {
		return fmt.Errorf("encrypt access token: %w", err)
	}
	encryptedRefreshToken, err := auth.AESGCMEncrypt(oauthToken.RefreshToken)
	if err != nil {
		return fmt.Errorf("encrypt refresh token: %w", err)
	}
	return gen.New(s.db).UserUpdateSpotifyTokens(ctx, gen.UserUpdateSpotifyTokensParams{
		UserID:                userUUID,
		EncryptedAccessToken:  encryptedAccessToken,
		AccessTokenExpiry:     oauthToken.Expiry,
		EncryptedRefreshToken: encryptedRefreshToken,
	})
}

func (s *Store) UpdateSpotifyInfo(ctx context.Context, userID string, info *spotify.SpotifyUser) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return gen.New(s.db).UserUpdateSpotifyInfo(ctx, gen.UserUpdateSpotifyInfoParams{
		ID:              userUUID,
		SpotifyAccount:  sql.NullString{String: info.ID, Valid: true},
		SpotifyName:     sql.NullString{String: info.Display, Valid: true},
		SpotifyImageUrl: sql.NullString{String: info.ImageURL, Valid: true},
	})
}

func (s *Store) Authenticate(ctx context.Context, username string, password string) (bool, error) {
	return gen.New(s.db).UserValidatePassword(ctx, gen.UserValidatePasswordParams{
		Username: username,
		UserPass: password,
	})
}

type GetUserRoomResponse struct {
	ID      string    `json:"id"`
	Code    string    `json:"code"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
}

func (s *Store) GetUserRoom(ctx context.Context, userID string) (*GetUserRoomResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user UUID: %w", err)
	}

	row, err := gen.New(s.db).UserGetRoom(ctx, userUUID)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &GetUserRoomResponse{
		ID:      row.ID.String(),
		Name:    row.Name,
		Code:    row.Code,
		Created: row.Created,
	}, nil
}

func (s *Store) GetByUsername(ctx context.Context, username string) (*User, error) {
	row, err := gen.New(s.db).UserGetByUsername(ctx, username)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, err
	}
	return &User{
		ID:           row.ID.String(),
		Username:     row.Username,
		DisplayName:  row.DisplayName,
		SpotifyName:  row.SpotifyName.String,
		SpotifyImage: row.SpotifyImageUrl.String,
	}, nil
}

func (s *Store) GetByID(ctx context.Context, userID string) (*User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user uuid: %s", err)
	}
	row, err := gen.New(s.db).UserGetByID(ctx, userUUID)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, err
	}
	return &User{
		ID:           row.ID.String(),
		Username:     row.Username,
		DisplayName:  row.DisplayName,
		SpotifyName:  row.SpotifyName.String,
		SpotifyImage: row.SpotifyImageUrl.String,
	}, nil
}

func (s *Store) UnlinkSpotify(ctx context.Context, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("parse user uuid: %s", err)
	}
	return gen.New(s.db).UserDeleteSpotifyInfo(ctx, userUUID)
}
