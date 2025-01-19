package room

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/user"
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

func GetByCode(ctx context.Context, dbtx db.DBTX, code string) (Room, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	row, err := db.New(dbtx).RoomGetByCode(ctx, strings.ToUpper(code))
	if err != nil {
		return Room{}, err
	}
	return Room{
		ID: row.ID.String(),
		Host: user.User{
			ID:           row.HostID.String(),
			Username:     row.HostUsername,
			DisplayName:  row.HostDisplay,
			SpotifyName:  util.StringFromPointer(row.HostSpotifyName),
			SpotifyImage: row.HostImage,
		},
		Code:    row.Code,
		Name:    row.Name,
		Created: row.Created,
	}, nil
}

func GetHostID(ctx context.Context, dbtx db.DBTX, code string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	id, err := db.New(dbtx).RoomGetHostID(ctx, strings.ToUpper(code))
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func GetEncryptedRoomTokens(ctx context.Context, dbtx db.DBTX, code string) (accessToken []byte, accessTokenExpiry time.Time, refreshToken []byte, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	row, err := db.New(dbtx).GetSpotifyTokensByRoomCode(ctx, strings.ToUpper(code))
	if err != nil {
		return
	}
	return row.EncryptedAccessToken, row.AccessTokenExpiry, row.EncryptedRefreshToken, nil
}

func Insert(ctx context.Context, dbtx db.DBTX, insertParams InsertRoomParams) (Room, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	hostUUID, err := uuid.Parse(insertParams.HostID)
	if err != nil {
		return Room{}, fmt.Errorf("parse user UUID: %w", err)
	}
	row, err := db.New(dbtx).RoomInsertWithPassword(
		ctx,
		db.RoomInsertWithPasswordParams{
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

func UpdateSpotifyToken(ctx context.Context, dbtx db.DBTX, code string, oauthToken *oauth2.Token) error {
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
	return db.New(dbtx).RoomUpdateSpotifyTokens(ctx, db.RoomUpdateSpotifyTokensParams{
		Code:                  strings.ToUpper(code),
		EncryptedAccessToken:  encryptedAccessToken,
		AccessTokenExpiry:     oauthToken.Expiry,
		EncryptedRefreshToken: encryptedRefreshToken,
	})
}

func ValidatePassword(ctx context.Context, dbtx db.DBTX, code string, password string) (bool, error) {
	return db.New(dbtx).RoomValidatePassword(ctx, db.RoomValidatePasswordParams{
		Code:     strings.ToUpper(code),
		RoomPass: password,
	})
}

func AddMember(ctx context.Context, dbtx db.DBTX, roomID string, userID string) error {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return fmt.Errorf("parse room UUID: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("parse user UUID: %w", err)
	}
	err = db.New(dbtx).RoomAddMember(ctx, db.RoomAddMemberParams{
		RoomID: roomUUID,
		UserID: userUUID,
	})
	return err
}

func AddMemberByUsername(ctx context.Context, dbtx db.DBTX, roomID string, username string, isModerator bool) error {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return fmt.Errorf("parse room UUID: %w", err)
	}
	_, err = db.New(dbtx).RoomAddMemberByUsername(ctx, db.RoomAddMemberByUsernameParams{
		RoomID:      roomUUID,
		Username:    username,
		IsModerator: isModerator,
	})
	return err
}

func UserIsMember(ctx context.Context, dbtx db.DBTX, roomID string, userID string) (bool, error) {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return false, fmt.Errorf("parse room UUID: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("parse room UUID: %w", err)
	}
	return db.New(dbtx).RoomUserIsMember(ctx, db.RoomUserIsMemberParams{UserID: userUUID, RoomID: roomUUID})
}

type Member struct {
	ID           string `json:"user_id"`
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	SpotifyName  string `json:"spotify_name"`
	SpotifyImage string `json:"spotify_image_url"`
	IsModerator  bool   `json:"is_moderator"`
	QueuedTracks int    `json:"queued_tracks"`
}

func GetAllMembers(ctx context.Context, dbtx db.DBTX, roomID string) ([]Member, error) {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return nil, fmt.Errorf("parse room UUID: %w", err)
	}
	rows, err := db.New(dbtx).RoomGetAllMembers(ctx, roomUUID)
	if err != nil {
		return nil, err
	}

	users := make([]Member, 0, len(rows))
	for _, row := range rows {
		users = append(users, Member{
			ID:           row.UserID.String(),
			Username:     row.Username,
			DisplayName:  row.DisplayName,
			SpotifyName:  util.StringFromPointer(row.SpotifyName),
			SpotifyImage: util.StringFromPointer(row.SpotifyImageUrl),
			IsModerator:  row.IsModerator,
			QueuedTracks: int(row.QueuedTracks),
		})
	}

	return users, nil
}

func SetModerator(ctx context.Context, dbtx db.DBTX, roomID string, userID string, isModerator bool) error {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return fmt.Errorf("parse room UUID: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("parse user UUID: %w", err)
	}
	err = db.New(dbtx).RoomSetModerator(ctx, db.RoomSetModeratorParams{
		RoomID:      roomUUID,
		UserID:      userUUID,
		IsModerator: isModerator,
	})
	return err
}

func RemoveMember(ctx context.Context, dbtx db.DBTX, roomID string, userID string) error {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return fmt.Errorf("parse room UUID: %w", err)
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("parse user UUID: %w", err)
	}
	return db.New(dbtx).RoomRemoveMember(ctx, db.RoomRemoveMemberParams{
		RoomID: roomUUID,
		UserID: userUUID,
	})
}

func InsertGuest(ctx context.Context, dbtx db.DBTX, roomCode string, name string) (*Guest, error) {
	row, err := db.New(dbtx).RoomGuestInsert(ctx, db.RoomGuestInsertParams{
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

func InsertGuestWithID(ctx context.Context, dbtx db.DBTX, roomCode string, name string, guestID string) (*Guest, error) {
	guestUUID, err := uuid.Parse(guestID)
	if err != nil {
		return nil, fmt.Errorf("parse guest UUID: %w", err)
	}

	row, err := db.New(dbtx).RoomGuestInsertWithID(ctx, db.RoomGuestInsertWithIDParams{
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

func GetGuestName(ctx context.Context, dbtx db.DBTX, roomID string, guestID string) (string, error) {
	guestUUID, err := uuid.Parse(guestID)
	if err != nil {
		return "", fmt.Errorf("parse guest UUID: %w", err)
	}
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return "", fmt.Errorf("parse room UUID: %w", err)
	}

	return db.New(dbtx).RoomGuestGetName(ctx, db.RoomGuestGetNameParams{
		GuestID: guestUUID,
		RoomID:  roomUUID,
	})
}

func GetAllRoomGuests(ctx context.Context, dbtx db.DBTX, roomID string) ([]Guest, error) {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return nil, fmt.Errorf("parse room UUID: %w", err)
	}
	rows, err := db.New(dbtx).RoomGetAllGuests(ctx, roomUUID)
	if err != nil {
		return nil, err
	}

	guests := make([]Guest, 0, len(rows))
	for _, row := range rows {
		guests = append(guests, Guest{
			ID:           row.ID.String(),
			Name:         row.Name,
			QueuedTracks: int(row.QueuedTracks),
		})
	}

	return guests, nil
}

func SetQueueTrackGuest(ctx context.Context, dbtx db.DBTX, roomCode string, trackID string, guestID string) error {
	guestUUID, err := uuid.Parse(guestID)
	if err != nil {
		return fmt.Errorf("parse guest id: %s", err)
	}

	return db.New(dbtx).RoomSetGuestQueueTrack(ctx, db.RoomSetGuestQueueTrackParams{
		GuestID:  guestUUID,
		RoomCode: roomCode,
		TrackID:  trackID,
	})
}

func SetQueueTrackUser(ctx context.Context, dbtx db.DBTX, roomCode string, trackID string, userID string) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("parse user id: %s", err)
	}

	return db.New(dbtx).RoomSetMemberQueueTrack(ctx, db.RoomSetMemberQueueTrackParams{
		UserID:   userUUID,
		RoomCode: roomCode,
		TrackID:  trackID,
	})
}

func GetQueueTrackAddedBy(ctx context.Context, dbtx db.DBTX, roomID string) (tracks []QueuedTrack, err error) {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return nil, fmt.Errorf("parse room UUID: %w", err)
	}
	rows, err := db.New(dbtx).RoomGetQueueTracks(ctx, roomUUID)
	if err != nil {
		return
	}

	for _, row := range rows {
		addedBy := ""
		if row.GuestName != nil {
			addedBy = *row.GuestName
		} else if row.MemberName != nil {
			addedBy = *row.MemberName
		}
		tracks = append(tracks, QueuedTrack{
			TrackID:   row.TrackID,
			AddedBy:   addedBy,
			Timestamp: row.Timestamp,
			Played:    row.Played,
		})
	}

	return
}

func DeleteByCode(ctx context.Context, dbtx db.DBTX, roomCode string) error {
	return db.New(dbtx).RoomDeleteByID(ctx, strings.ToUpper(roomCode))
}

func UpdatePassword(ctx context.Context, dbtx db.DBTX, roomID string, newPassword string) error {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return fmt.Errorf("parse room UUID: %w", err)
	}

	return db.New(dbtx).RoomUpdatePassword(ctx, db.RoomUpdatePasswordParams{
		RoomID:   roomUUID,
		RoomPass: newPassword,
	})
}

func GetUserHostedRooms(ctx context.Context, dbtx db.DBTX, userID string, isOpen bool) ([]Room, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user UUID: %w", err)
	}

	rows, err := db.New(dbtx).UserGetHostedRooms(ctx, db.UserGetHostedRoomsParams{ID: userUUID, IsOpen: isOpen})
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	rooms := make([]Room, len(rows))
	for i, row := range rows {
		rooms[i] = Room{
			ID:   row.ID.String(),
			Code: row.Code,
			Name: row.Name,
			Host: user.User{
				ID:           row.HostID.String(),
				Username:     row.HostUsername,
				DisplayName:  row.HostDisplayName,
				SpotifyImage: row.HostSpotifyImageUrl,
			},
			Created: row.Created,
		}
	}
	return rooms, nil
}

func GetUserJoinedRooms(ctx context.Context, dbtx db.DBTX, userID string, isOpen bool) ([]Room, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user UUID: %w", err)
	}

	rows, err := db.New(dbtx).UserGetJoinedRooms(ctx, db.UserGetJoinedRoomsParams{UserID: userUUID, IsOpen: isOpen})
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	rooms := make([]Room, len(rows))
	for i, row := range rows {
		rooms[i] = Room{
			ID:   row.ID.String(),
			Code: row.Code,
			Name: row.Name,
			Host: user.User{
				ID:           row.HostID.String(),
				Username:     row.HostUsername,
				DisplayName:  row.HostDisplayName,
				SpotifyImage: row.HostSpotifyImageUrl,
			},
			Created: row.Created,
		}
	}
	return rooms, nil
}

func SetIsOpen(ctx context.Context, dbtx db.DBTX, roomID string, isOpen bool) error {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return fmt.Errorf("parse room UUID: %w", err)
	}
	return db.New(dbtx).RoomSetIsOpen(ctx, db.RoomSetIsOpenParams{
		ID:     roomUUID,
		IsOpen: isOpen,
	})
}

func MarkTracksAsPlayedSince(ctx context.Context, dbtx db.DBTX, roomID string, since time.Time) error {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return fmt.Errorf("parse room UUID: %w", err)
	}

	return db.New(dbtx).RoomMarkTracksAsPlayed(ctx, db.RoomMarkTracksAsPlayedParams{
		RoomID:    roomUUID,
		Timestamp: since,
	})
}
