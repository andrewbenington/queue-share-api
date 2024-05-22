package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type TrackCacheInsertBulkNullableParams struct {
	ID           []string  `json:"id"`
	Uri          []string  `json:"uri"`
	Name         []string  `json:"name"`
	AlbumID      []string  `json:"album_id"`
	AlbumUri     []string  `json:"album_uri"`
	AlbumName    []string  `json:"album_name"`
	ArtistID     []string  `json:"artist_id"`
	ArtistUri    []string  `json:"artist_uri"`
	ArtistName   []string  `json:"artist_name"`
	ImageUrl     []*string `json:"image_url"`
	OtherArtists []*string `json:"other_artists"`
	DurationMs   []int32   `json:"duration_ms"`
	Popularity   []int32   `json:"popularity"`
	Explicit     []bool    `json:"explicit"`
	PreviewUrl   []string  `json:"preview_url"`
	DiscNumber   []int32   `json:"disc_number"`
	TrackNumber  []int32   `json:"track_number"`
	Type         []string  `json:"type"`
	ExternalIds  []*string `json:"external_ids"`
	Isrc         []*string `json:"isrc"`
}

func (q *Queries) TrackCacheInsertBulkNullable(ctx context.Context, arg TrackCacheInsertBulkNullableParams) error {
	_, err := q.db.ExecContext(ctx, trackCacheInsertBulk,
		pq.Array(arg.ID),
		pq.Array(arg.Uri),
		pq.Array(arg.Name),
		pq.Array(arg.AlbumID),
		pq.Array(arg.AlbumUri),
		pq.Array(arg.AlbumName),
		pq.Array(arg.ArtistID),
		pq.Array(arg.ArtistUri),
		pq.Array(arg.ArtistName),
		pq.Array(arg.ImageUrl),
		pq.Array(arg.OtherArtists),
		pq.Array(arg.DurationMs),
		pq.Array(arg.Popularity),
		pq.Array(arg.Explicit),
		pq.Array(arg.PreviewUrl),
		pq.Array(arg.DiscNumber),
		pq.Array(arg.TrackNumber),
		pq.Array(arg.Type),
		pq.Array(arg.ExternalIds),
		pq.Array(arg.Isrc),
	)
	return err
}

type AlbumCacheInsertBulkNullableParams struct {
	ID                   []string    `json:"id"`
	Uri                  []string    `json:"uri"`
	Name                 []string    `json:"name"`
	ArtistID             []string    `json:"artist_id"`
	ArtistUri            []string    `json:"artist_uri"`
	ArtistName           []string    `json:"artist_name"`
	AlbumGroup           []*string   `json:"album_group"`
	AlbumType            []*string   `json:"album_type"`
	ImageUrl             []*string   `json:"image_url"`
	ReleaseDate          []time.Time `json:"release_date"`
	ReleaseDatePrecision []*string   `json:"release_date_precision"`
	Genres               []*string   `json:"genres"`
	Popularity           []int32     `json:"popularity"`
}

func (q *Queries) AlbumCacheInsertBulkNullable(ctx context.Context, arg AlbumCacheInsertBulkNullableParams) error {
	_, err := q.db.ExecContext(ctx, albumCacheInsertBulk,
		pq.Array(arg.ID),
		pq.Array(arg.Uri),
		pq.Array(arg.Name),
		pq.Array(arg.ArtistID),
		pq.Array(arg.ArtistUri),
		pq.Array(arg.ArtistName),
		pq.Array(arg.AlbumGroup),
		pq.Array(arg.AlbumType),
		pq.Array(arg.ImageUrl),
		pq.Array(arg.ReleaseDate),
		pq.Array(arg.ReleaseDatePrecision),
		pq.Array(arg.Genres),
		pq.Array(arg.Popularity),
	)
	return err
}

type ArtistCacheInsertBulkNullableParams struct {
	ID            []string  `json:"id"`
	Uri           []string  `json:"uri"`
	Name          []string  `json:"name"`
	ImageUrl      []*string `json:"image_url"`
	Genres        []*string `json:"genres"`
	Popularity    []int32   `json:"popularity"`
	FollowerCount []int32   `json:"follower_count"`
}

func (q *Queries) ArtistCacheInsertBulkNullable(ctx context.Context, arg ArtistCacheInsertBulkNullableParams) error {
	_, err := q.db.ExecContext(ctx, artistCacheInsertBulk,
		pq.Array(arg.ID),
		pq.Array(arg.Uri),
		pq.Array(arg.Name),
		pq.Array(arg.ImageUrl),
		pq.Array(arg.Genres),
		pq.Array(arg.Popularity),
		pq.Array(arg.FollowerCount),
	)
	return err
}

type HistoryInsertBulkNullableParams struct {
	UserIds          []uuid.UUID `json:"user_ids"`
	Timestamp        []time.Time `json:"timestamp"`
	Platform         []string    `json:"platform"`
	MsPlayed         []int32     `json:"ms_played"`
	ConnCountry      []string    `json:"conn_country"`
	IpAddr           []string    `json:"ip_addr"`
	UserAgent        []string    `json:"user_agent"`
	TrackName        []string    `json:"track_name"`
	ArtistName       []string    `json:"artist_name"`
	AlbumName        []string    `json:"album_name"`
	SpotifyTrackUri  []string    `json:"spotify_track_uri"`
	SpotifyArtistUri []*string   `json:"spotify_artist_uri"`
	SpotifyAlbumUri  []*string   `json:"spotify_album_uri"`
	ReasonStart      []string    `json:"reason_start"`
	ReasonEnd        []string    `json:"reason_end"`
	Shuffle          []bool      `json:"shuffle"`
	Skipped          []bool      `json:"skipped"`
	Offline          []bool      `json:"offline"`
	OfflineTimestamp []time.Time `json:"offline_timestamp"`
	IncognitoMode    []bool      `json:"incognito_mode"`
	FromHistory      []bool      `json:"from_history"`
}

func (q *Queries) HistoryInsertBulkNullable(ctx context.Context, arg HistoryInsertBulkNullableParams) error {
	_, err := q.db.ExecContext(ctx, historyInsertBulk,
		pq.Array(arg.UserIds),
		pq.Array(arg.Timestamp),
		pq.Array(arg.Platform),
		pq.Array(arg.MsPlayed),
		pq.Array(arg.ConnCountry),
		pq.Array(arg.IpAddr),
		pq.Array(arg.UserAgent),
		pq.Array(arg.TrackName),
		pq.Array(arg.ArtistName),
		pq.Array(arg.AlbumName),
		pq.Array(arg.SpotifyTrackUri),
		pq.Array(arg.SpotifyArtistUri),
		pq.Array(arg.SpotifyAlbumUri),
		pq.Array(arg.ReasonStart),
		pq.Array(arg.ReasonEnd),
		pq.Array(arg.Shuffle),
		pq.Array(arg.Skipped),
		pq.Array(arg.Offline),
		pq.Array(arg.OfflineTimestamp),
		pq.Array(arg.IncognitoMode),
		pq.Array(arg.FromHistory),
	)
	return err
}
