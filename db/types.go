package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type TrackArtist struct {
	ID   string `json:"id"`
	URI  string `json:"uri"`
	Name string `json:"string"`
}

type TrackExternalIDs map[string]string

func (ids TrackExternalIDs) Value() (driver.Value, error) {
	if len(ids) == 0 {
		return "{}", nil
	}
	bytes, err := json.Marshal(ids)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

func (ids *TrackExternalIDs) Scan(src interface{}) (err error) {
	var extIDs TrackExternalIDs
	switch src := src.(type) {
	case string:
		err = json.Unmarshal([]byte(src), &extIDs)
	case []byte:
		err = json.Unmarshal(src, &extIDs)
	default:
		return errors.New("incompatible type for TrackExternalIDs")
	}
	if err != nil {
		return
	}
	*ids = extIDs
	return nil
}
