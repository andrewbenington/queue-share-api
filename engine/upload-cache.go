package engine

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/spotify"
	z_spotify "github.com/zmb3/spotify/v2"
	"golang.org/x/exp/maps"
)

func uploadTrackCache(ctx context.Context) error {
	cache := spotify.GetTrackCache()

	tracks := maps.Values(cache)

	log.Printf("There are %d tracks in the cache", len(tracks))

	transaction, err := db.Service().DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for i := range (len(tracks) / 100) + 1 {
		start := i * 100
		end := start + 100
		if end > len(tracks) {
			end = len(tracks)
		}

		params := insertParamsFromFullTracks(tracks[start:end])

		err = db.New(transaction).TrackCacheInsertBulkNullable(ctx, params)
		if err != nil {
			log.Println(err)
		}

		log.Printf("Uploaded %d/%d", end, len(tracks))
		time.Sleep(time.Millisecond)
	}

	err = transaction.Commit()
	if err != nil {
		return err
	}

	return nil
}

func uploadAlbumCache(ctx context.Context) error {
	cache := spotify.GetAlbumCache()

	albums := maps.Values(cache)

	log.Printf("There are %d albums in the cache", len(albums))

	transaction, err := db.Service().DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for i := range (len(albums) / 100) + 1 {
		start := i * 100
		end := start + 100
		if end > len(albums) {
			end = len(albums)
		}

		params := insertParamsFromFullAlbums(albums[start:end])

		err = db.New(transaction).AlbumCacheInsertBulkNullable(ctx, params)
		if err != nil {
			log.Println(err)
		}

		log.Printf("Uploaded %d/%d", end, len(albums))
		time.Sleep(time.Millisecond)
	}

	err = transaction.Commit()
	if err != nil {
		return err
	}

	return nil
}

func insertParamsFromFullTracks(tracks []z_spotify.FullTrack) db.TrackCacheInsertBulkNullableParams {
	params := db.TrackCacheInsertBulkNullableParams{
		ID:           []string{},
		Uri:          []string{},
		Name:         []string{},
		AlbumID:      []string{},
		AlbumUri:     []string{},
		AlbumName:    []string{},
		ArtistID:     []string{},
		ArtistUri:    []string{},
		ArtistName:   []string{},
		ImageUrl:     []*string{},
		OtherArtists: []*string{},
		DurationMs:   []int32{},
		Popularity:   []int32{},
		Explicit:     []bool{},
		PreviewUrl:   []string{},
		DiscNumber:   []int32{},
		TrackNumber:  []int32{},
		Type:         []string{},
		ExternalIds:  []*string{},
		Isrc:         []*string{},
	}

	for _, track := range tracks {
		album := track.Album
		artist := track.Artists[0]
		params.ID = append(params.ID, track.ID.String())
		params.Uri = append(params.Uri, string(track.URI))
		params.Name = append(params.Name, track.Name)
		params.AlbumID = append(params.AlbumID, album.ID.String())
		params.AlbumUri = append(params.AlbumUri, string(album.URI))
		params.AlbumName = append(params.AlbumName, album.Name)
		params.ArtistID = append(params.ArtistID, artist.ID.String())
		params.ArtistUri = append(params.ArtistUri, string(artist.URI))
		params.ArtistName = append(params.ArtistName, artist.Name)

		if len(track.Artists) > 1 {
			trackArtists := Map(track.Artists[1:], trackArtistFromSimple)
			bytes, err := json.Marshal(trackArtists)
			if err != nil {
				log.Println(err)
				params.OtherArtists = append(params.OtherArtists, nil)
			} else {
				str := string(bytes)
				params.OtherArtists = append(params.OtherArtists, &str)
			}
		} else {
			params.OtherArtists = append(params.OtherArtists, nil)
		}

		params.DurationMs = append(params.DurationMs, int32(track.Duration))
		params.Popularity = append(params.Popularity, int32(track.Popularity))
		params.Explicit = append(params.Explicit, track.Explicit)
		params.PreviewUrl = append(params.PreviewUrl, track.PreviewURL)
		params.DiscNumber = append(params.DiscNumber, int32(track.DiscNumber))
		params.TrackNumber = append(params.TrackNumber, int32(track.TrackNumber))
		params.Type = append(params.Type, track.Type)

		if track.ExternalIDs != nil {
			bytes, err := json.Marshal(track.ExternalIDs)
			if err != nil {
				log.Println(err)
				params.ExternalIds = append(params.ExternalIds, nil)
			} else {
				str := string(bytes)
				params.ExternalIds = append(params.ExternalIds, &str)
			}
		} else {
			params.ExternalIds = append(params.ExternalIds, nil)
		}

		if isrc, ok := track.ExternalIDs["isrc"]; ok {
			params.Isrc = append(params.Isrc, &isrc)
		} else {
			params.Isrc = append(params.Isrc, nil)
		}
	}

	return params
}

func trackArtistFromSimple(artist z_spotify.SimpleArtist) db.TrackArtist {
	return db.TrackArtist{
		Name: artist.Name,
		ID:   artist.ID.String(),
		URI:  string(artist.URI),
	}
}

func insertParamsFromFullAlbums(albums []z_spotify.FullAlbum) db.AlbumCacheInsertBulkNullableParams {
	params := db.AlbumCacheInsertBulkNullableParams{
		ID:                   []string{},
		Uri:                  []string{},
		Name:                 []string{},
		ArtistID:             []string{},
		ArtistUri:            []string{},
		ArtistName:           []string{},
		AlbumGroup:           []*string{},
		AlbumType:            []*string{},
		ImageUrl:             []*string{},
		ReleaseDate:          []time.Time{},
		ReleaseDatePrecision: []*string{},
		Genres:               []*string{},
		Popularity:           []int32{},
	}

	for _, album := range albums {
		artist := album.Artists[0]
		params.ID = append(params.ID, album.ID.String())
		params.Uri = append(params.Uri, string(album.URI))
		params.Name = append(params.Name, album.Name)
		params.ArtistID = append(params.ArtistID, artist.ID.String())
		params.ArtistUri = append(params.ArtistUri, string(artist.URI))
		params.ArtistName = append(params.ArtistName, artist.Name)
		params.AlbumGroup = append(params.AlbumGroup, &album.AlbumGroup)
		params.AlbumType = append(params.AlbumType, &album.AlbumType)
		image := spotify.Get64Image(album.SimpleAlbum)
		if image != nil {
			params.ImageUrl = append(params.ImageUrl, &image.URL)
		} else {
			params.ImageUrl = append(params.ImageUrl, nil)
		}
		params.ReleaseDate = append(params.ReleaseDate, album.ReleaseDateTime())
		params.ReleaseDatePrecision = append(params.ReleaseDatePrecision, &album.ReleaseDatePrecision)

		bytes, err := json.Marshal(album.Genres)
		if err != nil {
			log.Println(err)
			params.Genres = append(params.Genres, nil)
		} else {
			str := string(bytes)
			params.Genres = append(params.Genres, &str)
		}

		params.Popularity = append(params.Popularity, int32(album.Popularity))
	}

	return params
}

func Map[T any, K any](data []T, f func(elem T) K) []K {
	results := make([]K, len(data))

	for i, elem := range data {
		results[i] = f(elem)
	}

	return results
}
