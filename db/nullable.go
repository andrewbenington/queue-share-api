package db

import (
	"context"
	"time"

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
