package history

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/andrewbenington/queue-share-api/db/gen"
	"github.com/google/uuid"
)

type StreamingEntry struct {
	Timestamp        string  `json:"ts"`
	Username         string  `json:"username"`
	Platform         string  `json:"platform"`
	MsPlayed         int32   `json:"ms_played"`
	ConnCountry      string  `json:"conn_country"`
	IpAddr           *string `json:"ip_addr_decrypted"`
	UserAgent        *string `json:"user_agent_decrypted"`
	TrackName        string  `json:"master_metadata_track_name"`
	ArtistName       string  `json:"master_metadata_album_artist_name"`
	AlbumName        string  `json:"master_metadata_album_album_name"`
	SpotifyTrackUri  string  `json:"spotify_track_uri"`
	ReasonStart      *string `json:"reason_start"`
	ReasonEnd        *string `json:"reason_end"`
	Shuffle          bool    `json:"shuffle"`
	Skipped          *bool   `json:"skipped"`
	Offline          bool    `json:"offline"`
	OfflineTimestamp int64   `json:"offline_timestamp"`
	IncognitoMode    bool    `json:"incognito_mode"`
}

func InsertEntry(ctx context.Context, transaction gen.DBTX, entry gen.HistoryInsertOneParams) error {
	return gen.New(transaction).HistoryInsertOne(ctx, entry)
}

func InsertEntries(ctx context.Context, transaction gen.DBTX, userID uuid.UUID, entries []StreamingEntry) error {

	params := gen.HistoryInsertBulkParams{
		UserIds:          []uuid.UUID{},
		Timestamp:        []time.Time{},
		Platform:         []string{},
		MsPlayed:         []int32{},
		ConnCountry:      []string{},
		IpAddr:           []string{},
		UserAgent:        []string{},
		TrackName:        []string{},
		ArtistName:       []string{},
		AlbumName:        []string{},
		SpotifyTrackUri:  []string{},
		ReasonStart:      []string{},
		ReasonEnd:        []string{},
		Shuffle:          []bool{},
		Skipped:          []bool{},
		Offline:          []bool{},
		IncognitoMode:    []bool{},
		OfflineTimestamp: []time.Time{},
	}

	for _, entry := range entries {
		if entry.TrackName == "" {
			continue
		}

		params.UserIds = append(params.UserIds, userID)

		parsedTime, err := time.Parse("2006-01-02T15:04:05Z", entry.Timestamp)
		if err != nil {
			fmt.Println(err)
			continue
		}
		params.Timestamp = append(params.Timestamp, parsedTime)
		params.Platform = append(params.Platform, entry.Platform)
		params.MsPlayed = append(params.MsPlayed, entry.MsPlayed)
		params.ConnCountry = append(params.ConnCountry, entry.ConnCountry)
		params.IpAddr = append(params.IpAddr, NullStringFromPtr(entry.IpAddr).String)
		params.UserAgent = append(params.UserAgent, NullStringFromPtr(entry.UserAgent).String)
		params.TrackName = append(params.TrackName, entry.TrackName)
		params.ArtistName = append(params.ArtistName, entry.ArtistName)
		params.AlbumName = append(params.AlbumName, entry.AlbumName)
		params.SpotifyTrackUri = append(params.SpotifyTrackUri, entry.SpotifyTrackUri)
		params.ReasonStart = append(params.ReasonStart, NullStringFromPtr(entry.ReasonStart).String)
		params.ReasonEnd = append(params.ReasonEnd, NullStringFromPtr(entry.ReasonEnd).String)
		params.Shuffle = append(params.Shuffle, entry.Shuffle)
		params.Skipped = append(params.Skipped, NullBoolFromPtr(entry.Skipped).Bool)
		params.Offline = append(params.Offline, entry.Offline)
		params.IncognitoMode = append(params.IncognitoMode, entry.IncognitoMode)
		params.OfflineTimestamp = append(params.OfflineTimestamp, time.Unix(entry.OfflineTimestamp/1000, entry.OfflineTimestamp%1000))
	}
	return gen.New(transaction).HistoryInsertBulk(ctx, params)
}

func NullStringFromPtr(ptr *string) sql.NullString {
	if ptr == nil {
		return sql.NullString{}
	}
	return sql.NullString{Valid: true, String: *ptr}
}

func NullBoolFromPtr(ptr *bool) sql.NullBool {
	if ptr == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{Valid: true, Bool: *ptr}
}
