package engine

import (
	"context"
	"fmt"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/config"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/google/uuid"
)

// func loadURIsFromCache() {
// 	ctx := context.Background()
// 	cache := service.GetTrackCache()
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

var missingURICountByUser = map[string]int{}

func loadURIsByPopularity(ctx context.Context) int {
	total := 0

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		fmt.Printf("Could not connect to database to get missing artist URIs %s\n", err)
		return 0
	}
	defer tx.Commit(ctx)

	rows, err := db.New(tx).MissingArtistURIs(ctx)
	if err != nil {
		fmt.Printf("Could not get missing artist URIs %s\n", err)
		return 0
	}
	missingURICountByUser = map[string]int{}
	for _, row := range rows {
		for _, userID := range row.UserIds {
			if count, ok := missingURICountByUser[userID]; ok {
				missingURICountByUser[userID] = count + 1
			} else {
				missingURICountByUser[userID] = 1
			}
		}
	}

	for range 5 {
		rows, err := db.New(tx).HistoryGetTopTracksWithoutURIs(ctx)
		if err != nil {
			fmt.Printf("Could not get tracks without URIs: %s\n", err)
			return 0
		}

		var userID string
		if config.GetIsProd() {
			userID = "9123cddc-2b54-483b-ad86-119c332654c0"
		} else {
			userID = "8d0cc5d4-ab80-4395-b345-727dffdf9701"
		}

		_, spClient, err := client.ForUser(ctx, uuid.MustParse(userID))
		if err != nil {
			fmt.Printf("Could not get service client: %s\n", err)
			return 0
		}

		trackIDs := []string{}
		for _, row := range rows {
			id, err := service.IDFromURI(row.SpotifyTrackUri)
			if err != nil {
				fmt.Printf("Could not get service ID from uri %s\n", row.SpotifyTrackUri)
				continue
			}
			trackIDs = append(trackIDs, id)
		}

		tracks, err := service.GetTracks(ctx, spClient, trackIDs)
		if err != nil {
			fmt.Printf("Could not get track IDs %s\n", err)
			continue
		}
		total += len(tracks)

		for _, track := range tracks {
			db.New(tx).HistorySetURIsForTrack(ctx, db.HistorySetURIsForTrackParams{
				SpotifyTrackUri:  string(track.URI),
				SpotifyArtistUri: &track.ArtistURI,
				SpotifyAlbumUri:  &track.AlbumURI,
			})
			fmt.Printf("loaded URIs for %s\n", track.Name)
		}

	}

	rows, err = db.New(tx).MissingArtistURIs(ctx)
	if err != nil {
		fmt.Printf("Could not get missing artist URIs %s\n", err)
		return 0
	}
	missingURICountByUser = map[string]int{}
	for _, row := range rows {
		for _, userID := range row.UserIds {
			if count, ok := missingURICountByUser[userID]; ok {
				missingURICountByUser[userID] = count + 1
			} else {
				missingURICountByUser[userID] = 1
			}
		}
	}
	return total
}

func GetMissingURIsByUser() map[string]int {
	return missingURICountByUser
}

func cacheTracksByPopularity(ctx context.Context) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		fmt.Printf("Could not connect to database to get missing artist URIs %s\n", err)
		return
	}
	defer tx.Commit(ctx)

	for range 5 {
		rows, err := db.New(tx).HistoryGetTopTracksNotInCache(ctx)
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
			fmt.Printf("Could not get service client: %s\n", err)
			return
		}

		trackIDs := []string{}
		for _, row := range rows {
			id, err := service.IDFromURI(row.SpotifyTrackUri)
			if err != nil {
				fmt.Printf("Could not get service ID from uri %s\n", row.SpotifyTrackUri)
				continue
			}
			trackIDs = append(trackIDs, id)
		}

		tracks, err := service.GetTracks(ctx, spClient, trackIDs)
		if err != nil {
			fmt.Printf("Could not get track IDs %s\n", err)
			continue
		}

		fmt.Printf("Cached %d tracks\n", len(tracks))

	}
}

func cacheAlbumsByPopularity(ctx context.Context) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		fmt.Printf("Could not connect to database to get missing artist URIs %s\n", err)
		return
	}
	defer tx.Commit(ctx)

	for range 5 {
		rows, err := db.New(tx).HistoryGetTopAlbumsNotInCache(ctx)
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
			fmt.Printf("Could not get service client: %s\n", err)
			return
		}

		albumIDs := []string{}
		for _, row := range rows {
			id, err := service.IDFromURIPtr(row.SpotifyAlbumUri)
			if err != nil {
				if row.SpotifyAlbumUri != nil {
					fmt.Printf("Could not get service ID from uri %s\n", *row.SpotifyAlbumUri)
				} else {
					fmt.Println("Could not get service ID from uri")
				}
				continue
			}
			albumIDs = append(albumIDs, *id)
		}

		albums, err := service.GetAlbums(ctx, spClient, albumIDs)
		if err != nil {
			fmt.Printf("Could not get album IDs %s\n", err)
			continue
		}

		fmt.Printf("Cached %d albums\n", len(albums))

	}
}

func updateTrackImageByAlbumPopularity(ctx context.Context) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		fmt.Printf("Could not connect to database to get missing artist URIs %s\n", err)
		return
	}
	defer tx.Commit(ctx)

	for range 5 {
		rows, err := db.New(tx).HistoryGetTopAlbumsNotInCache(ctx)
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
			fmt.Printf("Could not get service client: %s\n", err)
			return
		}

		albumIDs := []string{}
		for _, row := range rows {
			id, err := service.IDFromURIPtr(row.SpotifyAlbumUri)
			if err != nil {
				if row.SpotifyAlbumUri != nil {
					fmt.Printf("Could not get service ID from uri %s\n", *row.SpotifyAlbumUri)
				} else {
					fmt.Println("Could not get service ID from uri")
				}
				continue
			}
			albumIDs = append(albumIDs, *id)
		}

		albums, err := service.GetAlbums(ctx, spClient, albumIDs)
		if err != nil {
			fmt.Printf("Could not get album IDs %s\n", err)
			continue
		}

		fmt.Printf("Cached %d albums\n", len(albums))

	}
}
