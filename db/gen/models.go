// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0

package gen

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID      uuid.UUID
	Name    string
	Code    string
	Created time.Time
	HostID  uuid.UUID
}

type RoomPassword struct {
	ID                uuid.UUID
	RoomID            uuid.UUID
	EncryptedPassword []byte
}

type SchemaMigration struct {
	Version int64
	Dirty   bool
}

type SpotifyToken struct {
	ID                    uuid.UUID
	UserID                uuid.UUID
	EncryptedAccessToken  []byte
	AccessTokenExpiry     time.Time
	EncryptedRefreshToken []byte
}

type User struct {
	ID      uuid.UUID
	Name    string
	Created time.Time
}

type UserPassword struct {
	ID                uuid.UUID
	UserID            uuid.NullUUID
	EncryptedPassword string
}
