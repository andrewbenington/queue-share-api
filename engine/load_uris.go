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

func loadURIsFromCache() {
	ctx := context.Background()
	cache := spotify.GetTrackCache()
	for _, track := range cache {
		firstArtistURI := sql.NullString{}

		if len(track.Artists) > 0 {
			firstArtistURI.String = string(track.Artists[0].URI)
			firstArtistURI.Valid = true
		}
		db.New(db.Service().DB).HistorySetURIsForTrack(ctx, db.HistorySetURIsForTrackParams{
			SpotifyTrackUri:  string(track.URI),
			SpotifyArtistUri: firstArtistURI,
			SpotifyAlbumUri:  sql.NullString{Valid: true, String: string(track.Album.URI)},
		})
		fmt.Printf("loaded URIs for %s\n", track.Name)
	}
}

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
			firstArtistURI := sql.NullString{}

			if len(track.Artists) > 0 {
				firstArtistURI.String = string(track.Artists[0].URI)
				firstArtistURI.Valid = true
			}
			db.New(db.Service().DB).HistorySetURIsForTrack(ctx, db.HistorySetURIsForTrackParams{
				SpotifyTrackUri:  string(track.URI),
				SpotifyArtistUri: firstArtistURI,
				SpotifyAlbumUri:  sql.NullString{Valid: true, String: string(track.Album.URI)},
			})
			fmt.Printf("loaded URIs for %s\n", track.Name)
		}

	}
}
