package room

import "time"

type Room struct {
	ID       string    `json:"id"`
	Code     string    `json:"code"`
	Name     string    `json:"name"`
	HostID   string    `json:"host_id"`
	HostName string    `json:"host_name"`
	Created  time.Time `json:"created"`
}
