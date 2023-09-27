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
