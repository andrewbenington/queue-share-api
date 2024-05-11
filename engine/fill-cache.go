package engine

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/config"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/google/uuid"
)

// func loadURIsFromCache() {
// 	ctx := context.Background()
// 	cache := spotify.GetTrackCache()
// 	for _, track := range cache {
// 		firstArtistURI := sql.NullString{}

// 		if len(track.Artists) > 0 {
// 			firstArtistURI.String = string(track.Artists[0].URI)
// 			firstArtistURI.Valid = true
// 		}
// 		db.New(db.Service().DB).HistorySetURIsForTrack(ctx, db.HistorySetURIsForTrackParams{
// 			SpotifyTrackUri:  string(track.URI),
// 			SpotifyArtistUri: firstArtistURI,
// 			SpotifyAlbumUri:  sql.NullString{Valid: true, String: string(track.Album.URI)},
// 		})
// 		fmt.Printf("loaded URIs for %s\n", track.Name)
// 	}
// }

func loadURIsByPopularity() {
	ctx := context.Background()
	for range 5 {
		rows, err := db.New(db.Service().DB).HistoryGetTopTracksWithoutURIs(ctx)
		if err != nil {
			fmt.Printf("Could not get tracks without URIs: %s\n", err)
			return
		}

		var userID string
		if config.GetIsProd() {
			userID = "9123cddc-2b54-483b-ad86-119c332654c0"
		} else {
			userID = "8d0cc5d4-ab80-4395-b345-727dffdf9701"
		}

		_, spClient, err := client.ForUser(ctx, uuid.MustParse(userID))
		if err != nil {
			fmt.Printf("Could not get spotify client: %s\n", err)
			return
		}

		trackIDs := []string{}
		for _, row := range rows {
			id, err := spotify.IDFromURI(row.SpotifyTrackUri)
			if err != nil {
				fmt.Printf("Could not get spotify ID from uri %s\n", row.SpotifyTrackUri)
				continue
			}
			trackIDs = append(trackIDs, id)
		}

		tracks, err := spotify.GetTracks(ctx, spClient, trackIDs)
		if err != nil {
			fmt.Printf("Could not get track IDs %s\n", err)
			continue
		}

		for _, track := range tracks {
			db.New(db.Service().DB).HistorySetURIsForTrack(ctx, db.HistorySetURIsForTrackParams{
				SpotifyTrackUri:  string(track.URI),
				SpotifyArtistUri: sql.NullString{Valid: true, String: string(track.ArtistURI)},
				SpotifyAlbumUri:  sql.NullString{Valid: true, String: string(track.AlbumURI)},
			})
			fmt.Printf("loaded URIs for %s\n", track.Name)
		}

	}
}

func cacheTracksByPopularity() {
	ctx := context.Background()
	for range 5 {
		rows, err := db.New(db.Service().DB).HistoryGetTopTracksNotInCache(ctx)
		if err != nil {
			fmt.Printf("Could not get tracks not in cache: %s\n", err)
			return
		}

		var userID string
		if config.GetIsProd() {
			userID = "9123cddc-2b54-483b-ad86-119c332654c0"
		} else {
			userID = "8d0cc5d4-ab80-4395-b345-727dffdf9701"
		}

		_, spClient, err := client.ForUser(ctx, uuid.MustParse(userID))
		if err != nil {
			fmt.Printf("Could not get spotify client: %s\n", err)
			return
		}

		trackIDs := []string{}
		for _, row := range rows {
			id, err := spotify.IDFromURI(row.SpotifyTrackUri)
			if err != nil {
				fmt.Printf("Could not get spotify ID from uri %s\n", row.SpotifyTrackUri)
				continue
			}
			trackIDs = append(trackIDs, id)
		}

		tracks, err := spotify.GetTracks(ctx, spClient, trackIDs)
		if err != nil {
			fmt.Printf("Could not get track IDs %s\n", err)
			continue
		}

		fmt.Printf("Cached %d tracks\n", len(tracks))

	}
}

func cacheAlbumsByPopularity() {
	ctx := context.Background()
	for range 5 {
		rows, err := db.New(db.Service().DB).HistoryGetTopAlbumsNotInCache(ctx)
		if err != nil {
			fmt.Printf("Could not get albums not in cache: %s\n", err)
			return
		}

		var userID string
		if config.GetIsProd() {
			userID = "9123cddc-2b54-483b-ad86-119c332654c0"
		} else {
			userID = "8d0cc5d4-ab80-4395-b345-727dffdf9701"
		}

		_, spClient, err := client.ForUser(ctx, uuid.MustParse(userID))
		if err != nil {
			fmt.Printf("Could not get spotify client: %s\n", err)
			return
		}

		albumIDs := []string{}
		for _, row := range rows {
			id, err := spotify.IDFromURI(row.SpotifyAlbumUri.String)
			if err != nil {
				fmt.Printf("Could not get spotify ID from uri %s\n", row.SpotifyAlbumUri.String)
				continue
			}
			albumIDs = append(albumIDs, id)
		}

		albums, err := spotify.GetAlbums(ctx, spClient, albumIDs)
		if err != nil {
			fmt.Printf("Could not get album IDs %s\n", err)
			continue
		}

		fmt.Printf("Cached %d albums\n", len(albums))

	}
}

func updateTrackImageByAlbumPopularity() {
	ctx := context.Background()
	for range 5 {
		rows, err := db.New(db.Service().DB).HistoryGetTopAlbumsNotInCache(ctx)
		if err != nil {
			fmt.Printf("Could not get albums not in cache: %s\n", err)
			return
		}

		var userID string
		if config.GetIsProd() {
			userID = "9123cddc-2b54-483b-ad86-119c332654c0"
		} else {
			userID = "8d0cc5d4-ab80-4395-b345-727dffdf9701"
		}

		_, spClient, err := client.ForUser(ctx, uuid.MustParse(userID))
		if err != nil {
			fmt.Printf("Could not get spotify client: %s\n", err)
			return
		}

		albumIDs := []string{}
		for _, row := range rows {
			id, err := spotify.IDFromURI(row.SpotifyAlbumUri.String)
			if err != nil {
				fmt.Printf("Could not get spotify ID from uri %s\n", row.SpotifyAlbumUri.String)
				continue
			}
			albumIDs = append(albumIDs, id)
		}

		albums, err := spotify.GetAlbums(ctx, spClient, albumIDs)
		if err != nil {
			fmt.Printf("Could not get album IDs %s\n", err)
			continue
		}

		fmt.Printf("Cached %d albums\n", len(albums))

	}
}
