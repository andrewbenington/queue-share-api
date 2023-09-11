package room

import "time"

type Room struct {
	ID      string    `json:"id"`
	Code    string    `json:"code"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
}
