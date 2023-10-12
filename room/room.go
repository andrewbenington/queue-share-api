package room

import (
	"time"

	"github.com/andrewbenington/queue-share-api/user"
)

type Room struct {
	ID      string    `json:"id"`
	Code    string    `json:"code"`
	Name    string    `json:"name"`
	Host    user.User `json:"host"`
	Created time.Time `json:"created"`
}

type InsertRoomParams struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	HostID   string `json:"host_id"`
}

type InsertGuestRequest struct {
	Name string
}

type Guest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RoomResponse struct {
	Room  Room   `json:"room"`
	Guest *Guest `json:"guest_data"`
}

type TrackWithGuest struct {
	TrackID   string
	GuestName string
}
