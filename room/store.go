package room

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/db/gen"
	"github.com/andrewbenington/queue-share-api/user"
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
	row, err := gen.New(s.db).FindRoomByCode(ctx, strings.ToUpper(code))
	if err != nil {
		return Room{}, err
	}
	return Room{
		ID: row.ID.String(),
		Host: user.User{
			ID:           row.HostID.String(),
			Username:     row.HostUsername,
			DisplayName:  row.HostDisplay,
			SpotifyName:  row.HostSpotifyName.String,
			SpotifyImage: row.HostImage.String,
		},
		Code:    row.Code,
		Name:    row.Name,
		Created: row.Created,
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

func (s *Store) Insert(ctx context.Context, insertParams InsertRoomParams) (Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	hostUUID, err := uuid.Parse(insertParams.HostID)
	if err != nil {
		return Room{}, fmt.Errorf("parse user UUID: %w", err)
	}
	row, err := gen.New(s.db).InsertRoomWithPass(
		ctx,
		gen.InsertRoomWithPassParams{
			Name:     insertParams.Name,
			HostID:   hostUUID,
			RoomPass: insertParams.Password,
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
		Code:                  strings.ToUpper(code),
		EncryptedAccessToken:  encryptedAccessToken,
		AccessTokenExpiry:     oauthToken.Expiry,
		EncryptedRefreshToken: encryptedRefreshToken,
	})
}

func (s *Store) ValidatePassword(ctx context.Context, code string, password string) (bool, error) {
	return gen.New(s.db).ValidateRoomPass(ctx, gen.ValidateRoomPassParams{
		Code:     strings.ToUpper(code),
		RoomPass: password,
	})
}

func (s *Store) InsertGuest(ctx context.Context, roomCode string, name string) (*Guest, error) {
	row, err := gen.New(s.db).RoomGuestInsert(ctx, gen.RoomGuestInsertParams{
		RoomCode: roomCode,
		Name:     name,
	})
	if err != nil {
		return nil, err
	}
	return &Guest{
		ID:   row.ID.String(),
		Name: row.Name,
	}, nil
}

func (s *Store) InsertGuestWithID(ctx context.Context, roomCode string, name string, guestID string) (*Guest, error) {
	guestUUID, err := uuid.Parse(guestID)
	if err != nil {
		return nil, fmt.Errorf("parse guest UUID: %w", err)
	}

	row, err := gen.New(s.db).RoomGuestInsertWithID(ctx, gen.RoomGuestInsertWithIDParams{
		GuestID:  guestUUID,
		RoomCode: roomCode,
		Name:     name,
	})
	if err != nil {
		return nil, err
	}
	return &Guest{
		ID:   row.ID.String(),
		Name: row.Name,
	}, nil
}

func (s *Store) GetGuestName(ctx context.Context, roomID string, guestID string) (string, error) {
	guestUUID, err := uuid.Parse(guestID)
	if err != nil {
		return "", fmt.Errorf("parse guest UUID: %w", err)
	}
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return "", fmt.Errorf("parse guest UUID: %w", err)
	}

	return gen.New(s.db).RoomGuestGetName(ctx, gen.RoomGuestGetNameParams{
		GuestID: guestUUID,
		RoomID:  roomUUID,
	})
}

func (s *Store) SetQueueTrackGuest(ctx context.Context, roomCode string, trackID string, guestID string) error {
	guestUUID, err := uuid.Parse(guestID)
	if err != nil {
		return fmt.Errorf("parse guest id: %s", err)
	}

	return gen.New(s.db).RoomSetQueueTrack(ctx, gen.RoomSetQueueTrackParams{
		GuestID:  guestUUID,
		RoomCode: roomCode,
		TrackID:  trackID,
	})
}

func (s *Store) GetQueueTrackGuests(ctx context.Context, roomCode string) (tracks []TrackWithGuest, err error) {
	rows, err := gen.New(s.db).RoomGetQueueTracks(ctx, roomCode)
	if err != nil {
		return
	}

	for _, row := range rows {
		tracks = append(tracks, TrackWithGuest{
			TrackID:   row.TrackID,
			GuestName: row.Name,
		})
	}

	return
}
