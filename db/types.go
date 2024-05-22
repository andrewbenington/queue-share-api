package db

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type TrackArtist struct {
	ID   string `json:"id"`
	URI  string `json:"uri"`
	Name string `json:"name"`
}

type TrackArtists []TrackArtist

func (a TrackArtists) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	bytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

func (a *TrackArtists) Scan(src interface{}) (err error) {
	if src == nil {
		return nil
	}

	var artists TrackArtists
	switch src := src.(type) {
	case string:
		err = json.Unmarshal([]byte(src), &artists)
	case []byte:
		err = json.Unmarshal(src, &artists)
	default:
		return errors.New("incompatible type for TrackArtists")
	}
	if err != nil {
		return
	}
	*a = artists
	return nil
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
