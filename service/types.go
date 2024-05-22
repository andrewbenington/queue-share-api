package service

import "github.com/zmb3/spotify/v2"

type TrackData struct {
	ID           string                   `json:"id"`
	Uri          string                   `json:"uri"`
	Name         string                   `json:"name"`
	AlbumID      string                   `json:"album_id"`
	AlbumUri     string                   `json:"album_uri"`
	AlbumName    string                   `json:"album_name"`
	ArtistID     string                   `json:"artist_id"`
	ArtistUri    string                   `json:"artist_uri"`
	ArtistName   string                   `json:"artist_name"`
	ImageUrl     *string                  `json:"image_url"`
	OtherArtists TrackArtist              `json:"other_artists"`
	DurationMs   int32                    `json:"duration_ms"`
	Popularity   int32                    `json:"popularity"`
	Explicit     bool                     `json:"explicit"`
	PreviewUrl   string                   `json:"preview_url"`
	DiscNumber   int32                    `json:"disc_number"`
	TrackNumber  int32                    `json:"track_number"`
	Type         string                   `json:"type"`
	ExternalIds  spotify.TrackExternalIDs `json:"external_ids"`
	Isrc         *string                  `json:"isrc"`
}

type TrackArtist struct {
	ID   string `json:"id"`
	URI  string `json:"uri"`
	Name string `json:"name"`
}
