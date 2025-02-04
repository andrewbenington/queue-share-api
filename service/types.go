package service

import (
	"fmt"
	"time"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
)

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

type SimpleAlbumData struct {
	ID                   string     `json:"id"`
	URI                  string     `json:"uri"`
	Name                 string     `json:"name"`
	ArtistID             string     `json:"artist_id"`
	ArtistURI            string     `json:"artist_uri"`
	ArtistName           string     `json:"artist_name"`
	AlbumGroup           *string    `json:"album_group"`
	AlbumType            *string    `json:"album_type"`
	ImageUrl             *string    `json:"image_url"`
	ReleaseDate          *time.Time `json:"release_date"`
	ReleaseDatePrecision *string    `json:"release_date_precision"`
}

func TrackDataFromFullTrackIdx(ft spotify.FullTrack, _ int) db.TrackData {
	return TrackDataFromFullTrack(ft)
}

func TrackDataFromFullTrack(ft spotify.FullTrack) db.TrackData {
	artist := ft.Artists[0]

	image := GetAlbum300Image(ft.Album)
	var imageURL *string
	if image != nil {
		imageURL = &image.URL
	}

	var isrc *string
	if isrcField, ok := ft.ExternalIDs["isrc"]; ok {
		isrc = &isrcField
	}

	return db.TrackData{
		ID:           ft.ID.String(),
		URI:          string(ft.URI),
		Name:         ft.Name,
		AlbumID:      ft.Album.ID.String(),
		AlbumURI:     string(ft.Album.URI),
		AlbumName:    ft.Album.Name,
		ArtistID:     artist.ID.String(),
		ArtistURI:    string(artist.URI),
		ArtistName:   artist.Name,
		ImageUrl:     imageURL,
		OtherArtists: lo.Map(ft.Artists[1:], trackArtistFromSimple),
		DurationMs:   int32(ft.Duration),
		Popularity:   int32(ft.Popularity),
		Explicit:     ft.Explicit,
		PreviewUrl:   ft.PreviewURL,
		DiscNumber:   int32(ft.DiscNumber),
		TrackNumber:  int32(ft.TrackNumber),
		Type:         ft.Type,
		ExternalIds:  ft.ExternalIDs,
		Isrc:         isrc,
	}
}

func ArtistDataFromFullArtistIdx(fa spotify.FullArtist, _ int) db.ArtistData {
	return ArtistDataFromFullArtist(fa)
}

func ArtistDataFromFullArtist(fa spotify.FullArtist) db.ArtistData {
	image := GetArtist300Image(fa)
	var imageURL *string
	if image != nil {
		imageURL = &image.URL
	}

	popularity := int32(fa.Popularity)
	followerCount := int32(fa.Followers.Count)

	return db.ArtistData{
		ID:            fa.ID.String(),
		URI:           string(fa.URI),
		Name:          fa.Name,
		ImageUrl:      imageURL,
		Genres:        fa.Genres,
		Popularity:    &popularity,
		FollowerCount: &followerCount,
	}
}

func nilIfZeroYear(t time.Time) *time.Time {
	if t.Year() == 0 {
		return nil
	}
	return &t
}
func SimpleAlbumDataFromSimpleAlbum(a spotify.SimpleAlbum) (SimpleAlbumData, error) {
	if len(a.Artists) == 0 {
		return SimpleAlbumData{}, fmt.Errorf("album %s (%s) has no artists", a.Name, a.ID.String())
	}

	image := GetSimpleAlbum128Image(a)
	var imageURL *string
	if image != nil {
		imageURL = &image.URL
	}

	return SimpleAlbumData{
		ID:                   a.ID.String(),
		URI:                  string(a.URI),
		Name:                 a.Name,
		ArtistID:             a.Artists[0].ID.String(),
		ArtistURI:            string(a.Artists[0].URI),
		ArtistName:           a.Artists[0].Name,
		AlbumGroup:           &a.AlbumGroup,
		AlbumType:            &a.AlbumType,
		ImageUrl:             imageURL,
		ReleaseDate:          nilIfZeroYear(a.ReleaseDateTime()),
		ReleaseDatePrecision: &a.ReleaseDatePrecision,
	}, nil
}

func AlbumDataFromSpotifyAlbum(a spotify.FullAlbum) (*db.AlbumData, error) {
	if len(a.Artists) == 0 {
		return nil, fmt.Errorf("album %s (%s) has no artists", a.Name, a.ID.String())
	}

	trackIDs := make([]string, 0, len(a.Tracks.Tracks))
	trackISRCs := make([]string, 0, len(a.Tracks.Tracks))
	for _, track := range a.Tracks.Tracks {
		trackIDs = append(trackIDs, track.ID.String())
		if track.ExternalIDs.ISRC != "" {
			trackISRCs = append(trackISRCs, track.ExternalIDs.ISRC)
		}
	}

	image := GetAlbum128Image(a)
	var imageURL *string
	if image != nil {
		imageURL = &image.URL
	}
	popularity := int32(a.Popularity)

	var upc *string
	if val, ok := a.ExternalIDs["upc"]; ok {
		upc = &val
	}

	return &db.AlbumData{
		ID:                   a.ID.String(),
		URI:                  string(a.URI),
		Name:                 a.Name,
		ArtistID:             a.Artists[0].ID.String(),
		ArtistURI:            string(a.Artists[0].URI),
		ArtistName:           a.Artists[0].Name,
		AlbumGroup:           &a.AlbumGroup,
		AlbumType:            &a.AlbumType,
		ImageUrl:             imageURL,
		ReleaseDate:          nilIfZeroYear(a.ReleaseDateTime()),
		ReleaseDatePrecision: &a.ReleaseDatePrecision,
		Genres:               a.Genres,
		Popularity:           &popularity,
		Upc:                  upc,
		SpotifyTrackIds:      trackIDs,
		TrackIsrcs:           trackISRCs,
	}, nil
}
