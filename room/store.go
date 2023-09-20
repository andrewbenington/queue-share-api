package room

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/db/gen"
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

func (s *Store) GetByCode(ctx context.Context, code string) (Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	row, err := gen.New(s.db).FindRoomByCode(ctx, code)
	if err != nil {
		return Room{}, err
	}
	return Room{
		ID:       row.ID.String(),
		HostID:   row.HostID.String(),
		HostName: row.HostName,
		Code:     row.Code,
		Name:     row.Name,
		Created:  row.Created,
	}, nil
}

func (s *Store) GetEncryptedRoomTokens(ctx context.Context, code string) (accessToken []byte, accessTokenExpiry time.Time, refreshToken []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	row, err := gen.New(s.db).GetSpotifyTokensByRoomCode(ctx, code)
	if err != nil {
		return
	}
	return row.EncryptedAccessToken, row.AccessTokenExpiry, row.EncryptedRefreshToken, nil
}

type InsertRoomParams struct {
	Name              string    `json:"name"`
	Password          string    `json:"password"`
	HostName          string    `json:"host_name"`
	HostID            string    `json:"host_id"`
	AccessToken       string    `json:"access_token"`
	AccessTokenExpiry time.Time `json:"access_token_expiry"`
	RefreshToken      string    `json:"refresh_token"`
}

func (s *Store) Insert(ctx context.Context, insertParams InsertRoomParams) (Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	hostUUID, err := uuid.Parse(insertParams.HostID)
	if err != nil {
		return Room{}, fmt.Errorf("parse user UUID: %w", err)
	}
	row, err := gen.New(s.db).InsertRoom(
		ctx,
		gen.InsertRoomParams{
			Name:   insertParams.Name,
			HostID: hostUUID,
		},
	)
	if err != nil {
		return Room{}, err
	}
	return Room{
		ID:      row.ID.String(),
		Code:    row.Code,
		Name:    row.Name,
		Created: row.Created,
	}, nil
}

func (s *Store) UpdateSpotifyToken(ctx context.Context, code string, oauthToken *oauth2.Token) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	encryptedAccessToken, err := auth.AESGCMEncrypt(oauthToken.AccessToken)
	if err != nil {
		return fmt.Errorf("encrypt access token: %w", err)
	}
	encryptedRefreshToken, err := auth.AESGCMEncrypt(oauthToken.RefreshToken)
	if err != nil {
		return fmt.Errorf("encrypt refresh token: %w", err)
	}
	return gen.New(s.db).UpdateSpotifyTokensByRoomCode(ctx, gen.UpdateSpotifyTokensByRoomCodeParams{
		Code:                  code,
		EncryptedAccessToken:  encryptedAccessToken,
		AccessTokenExpiry:     oauthToken.Expiry,
		EncryptedRefreshToken: encryptedRefreshToken,
	})
}
