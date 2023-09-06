package room

import "time"

type Room struct {
	ID      string
	Code    string
	Name    string
	Created time.Time
}
