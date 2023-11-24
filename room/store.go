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

func (s *Store) GetHostID(ctx context.Context, code string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	id, err := gen.New(s.db).RoomGetHostID(ctx, strings.ToUpper(code))
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (s *Store) GetEncryptedRoomTokens(ctx context.Context, code string) (accessToken []byte, accessTokenExpiry time.Time, refreshToken []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	row, err := gen.New(s.db).GetSpotifyTokensByRoomCode(ctx, strings.ToUpper(code))
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

func (s *Store) AddMember(ctx context.Context, roomCode string, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("parse user UUID: %w", err)
	}
	err = gen.New(s.db).RoomAddMember(ctx, gen.RoomAddMemberParams{
		RoomCode: roomCode,
		UserID:   userUUID,
	})
	return err
}

func (s *Store) UserIsMember(ctx context.Context, roomID string, userID string) (bool, error) {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return false, fmt.Errorf("parse room UUID: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("parse room UUID: %w", err)
	}
	return gen.New(s.db).RoomUserIsMember(ctx, gen.RoomUserIsMemberParams{UserID: userUUID, RoomID: roomUUID})
}

func (s *Store) GetAllMembers(ctx context.Context, roomID string) ([]user.User, error) {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return nil, fmt.Errorf("parse room UUID: %w", err)
	}
	rows, err := gen.New(s.db).RoomGetAllMembers(ctx, roomUUID)
	if err != nil {
		return nil, err
	}

	users := make([]user.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, user.User{
			ID:           row.UserID.String(),
			Username:     row.Username,
			DisplayName:  row.DisplayName,
			SpotifyName:  row.SpotifyName.String,
			SpotifyImage: row.SpotifyImageUrl.String,
		})
	}

	return users, nil
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

func (s *Store) GetAllRoomGuests(ctx context.Context, roomCode string) ([]Guest, error) {
	rows, err := gen.New(s.db).RoomGetAllGuests(ctx, roomCode)
	if err != nil {
		return nil, err
	}

	guests := make([]Guest, 0, len(rows))
	for _, row := range rows {
		guests = append(guests, Guest{
			ID:   row.ID.String(),
			Name: row.Name,
		})
	}

	return guests, nil
}

func (s *Store) SetQueueTrackGuest(ctx context.Context, roomCode string, trackID string, guestID string) error {
	guestUUID, err := uuid.Parse(guestID)
	if err != nil {
		return fmt.Errorf("parse guest id: %s", err)
	}

	return gen.New(s.db).RoomSetGuestQueueTrack(ctx, gen.RoomSetGuestQueueTrackParams{
		GuestID:  guestUUID,
		RoomCode: roomCode,
		TrackID:  trackID,
	})
}

func (s *Store) SetQueueTrackUser(ctx context.Context, roomCode string, trackID string, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("parse user id: %s", err)
	}

	return gen.New(s.db).RoomSetMemberQueueTrack(ctx, gen.RoomSetMemberQueueTrackParams{
		UserID:   userUUID,
		RoomCode: roomCode,
		TrackID:  trackID,
	})
}

func (s *Store) GetQueueTrackGuests(ctx context.Context, roomCode string) (tracks []QueuedTrack, err error) {
	rows, err := gen.New(s.db).RoomGetQueueTracks(ctx, strings.ToUpper(roomCode))
	if err != nil {
		return
	}

	for _, row := range rows {
		addedBy := ""
		if row.GuestName.Valid {
			addedBy = row.GuestName.String
		} else if row.MemberName.Valid {
			addedBy = row.MemberName.String
		}
		tracks = append(tracks, QueuedTrack{
			TrackID: row.TrackID,
			AddedBy: addedBy,
		})
	}

	return
}

func (s *Store) DeleteByCode(ctx context.Context, roomCode string) error {
	return gen.New(s.db).RoomDeleteByID(ctx, strings.ToUpper(roomCode))
}
