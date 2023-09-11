package room

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/andrewbenington/go-spotify/auth"
	"github.com/andrewbenington/go-spotify/db/gen"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) All(ctx context.Context) ([]Room, error) {
	rows, err := gen.New(s.db).AllRooms(ctx)
	if err != nil {
		return nil, err
	}
	rooms := make([]Room, len(rows))
	for i, row := range rows {
		rooms[i] = Room{
			ID:      row.ID.String(),
			Code:    row.Code,
			Name:    row.Name,
			Created: row.Created.Time,
		}
	}
	return rooms, nil
}

func (s *Store) GetByCode(ctx context.Context, code string) (Room, error) {
	row, err := gen.New(s.db).FindRoomByCode(ctx, code)
	if err != nil {
		return Room{}, err
	}
	return Room{
		ID:      row.ID.String(),
		Code:    row.Code,
		Name:    row.Name,
		Created: row.Created.Time,
	}, nil
}

func (s *Store) GetRoomExistsByCode(ctx context.Context, code string) (bool, error) {
	_, err := gen.New(s.db).GetRoomIDByCode(ctx, code)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func (s *Store) GetEncryptedRoomTokens(ctx context.Context, code string) (accessToken string, accessTokenExpiry time.Time, refreshToken string, err error) {
	row, err := gen.New(s.db).GetRoomAuthByCode(ctx, code)
	if err != nil {
		return
	}
	return row.EncryptedAccessToken, row.AccessTokenExpiry, row.EncryptedRefreshToken, nil
}

type InsertRoomParams struct {
	Name              string    `json:"name"`
	Password          string    `json:"password"`
	AccessToken       string    `json:"access_token"`
	AccessTokenExpiry time.Time `json:"access_token_expiry"`
	RefreshToken      string    `json:"refresh_token"`
}

func (s *Store) Insert(ctx context.Context, insertParams InsertRoomParams) (Room, error) {
	encryptedAccessToken, err := auth.EncryptToken(insertParams.AccessToken, insertParams.Password)
	if err != nil {
		return Room{}, fmt.Errorf("encrypt access token with password: %w", err)
	}
	encryptedRefreshToken, err := auth.EncryptToken(insertParams.RefreshToken, insertParams.Password)
	if err != nil {
		return Room{}, fmt.Errorf("encrypt refresh token with password: %w", err)
	}
	row, err := gen.New(s.db).InsertRoom(
		ctx,
		gen.InsertRoomParams{
			Name:                  insertParams.Name,
			EncryptedAccessToken:  encryptedAccessToken,
			AccessTokenExpiry:     insertParams.AccessTokenExpiry,
			EncryptedRefreshToken: encryptedRefreshToken,
		},
	)
	if err != nil {
		return Room{}, err
	}
	return Room{
		ID:      row.ID.String(),
		Code:    row.Code,
		Name:    row.Name,
		Created: row.Created.Time,
	}, nil
}
