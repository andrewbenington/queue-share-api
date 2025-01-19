package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/andrewbenington/queue-share-api/util"
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

func InsertUser(ctx context.Context, dbtx db.DBTX, username string, displayName string, password string) (newUserID string, err error) {
	userUUID, err := db.New(dbtx).UserInsertWithPassword(ctx, db.UserInsertWithPasswordParams{
		Username:    username,
		DisplayName: displayName,
		UserPass:    password,
	})
	if err != nil {
		return "", err
	}
	return userUUID.String(), nil
}

func UpdateSpotifyToken(ctx context.Context, dbtx db.DBTX, userID string, oauthToken *oauth2.Token) error {
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
	return db.New(dbtx).UserUpdateSpotifyTokens(ctx, db.UserUpdateSpotifyTokensParams{
		UserID:                userUUID,
		EncryptedAccessToken:  encryptedAccessToken,
		AccessTokenExpiry:     oauthToken.Expiry,
		EncryptedRefreshToken: encryptedRefreshToken,
	})
}

func UpdateSpotifyInfo(ctx context.Context, dbtx db.DBTX, userID string, info *service.SpotifyUser) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return db.New(dbtx).UserUpdateSpotifyInfo(ctx, db.UserUpdateSpotifyInfoParams{
		ID:              userUUID,
		SpotifyAccount:  &info.ID,
		SpotifyName:     &info.Display,
		SpotifyImageUrl: &info.ImageURL,
	})
}

func Authenticate(ctx context.Context, dbtx db.DBTX, username string, password string) (bool, error) {
	return db.New(dbtx).UserValidatePassword(ctx, db.UserValidatePasswordParams{
		Username: username,
		UserPass: password,
	})
}

type HostedRoom struct {
	ID      string    `json:"id"`
	Code    string    `json:"code"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
}

func GetByUsername(ctx context.Context, dbtx db.DBTX, username string) (*User, error) {
	row, err := db.New(dbtx).UserGetByUsername(ctx, username)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	return &User{
		ID:           row.ID.String(),
		Username:     row.Username,
		DisplayName:  row.DisplayName,
		SpotifyName:  util.StringFromPointer(row.SpotifyName),
		SpotifyImage: row.SpotifyImageUrl,
	}, nil
}

func GetByID(ctx context.Context, dbtx db.DBTX, userID string) (*User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user uuid: %s", err)
	}
	row, err := db.New(dbtx).UserGetByID(ctx, userUUID)
	if err == sql.ErrNoRows {
		return nil, err
	} else if err != nil {
		return nil, err
	}
	return &User{
		ID:           row.ID.String(),
		Username:     row.Username,
		DisplayName:  row.DisplayName,
		SpotifyName:  util.StringFromPointer(row.SpotifyName),
		SpotifyImage: row.SpotifyImageUrl,
	}, nil
}

func UnlinkSpotify(ctx context.Context, dbtx db.DBTX, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("parse user uuid: %w", err)
	}
	err = db.New(dbtx).UserDeleteSpotifyInfo(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("delete user spotify data: %w", err)
	}
	err = db.New(dbtx).UserDeleteSpotifyToken(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("delete user spotify token: %w", err)
	}
	return nil
}
