package engine

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/google/uuid"
	z_spotify "github.com/zmb3/spotify/v2"
)

var (
	LastFetch *time.Time
)

func doHistoryCycle(ctx context.Context) {
	userIDs := []string{
		"9123cddc-2b54-483b-ad86-119c332654c0", // aab
		"531f7a17-3e0a-47b3-bb2c-4425f0150363", // al
		"ba7a624a-3352-4481-aaa5-d89f7c02a876", // ssb
		"0e868ac1-6e66-4745-9440-390d8241e92a", // mad
	}
	for _, userID := range userIDs {
		cancelCtx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		getHistoryForUser(cancelCtx, uuid.MustParse(userID))
		log.Printf("History fetched for user %s\n", userID)
	}
	now := time.Now()
	LastFetch = &now
}

func getHistoryForUser(ctx context.Context, userID uuid.UUID) {
	_, spClient, err := client.ForUser(ctx, userID)
	if err != nil {
		fmt.Println(err)
		return
	}

	var rows *[]z_spotify.RecentlyPlayedItem
	var before *time.Time

	for rows == nil || len(*rows) > 0 {
		opts := z_spotify.RecentlyPlayedOptions{Limit: 50}
		if before != nil {
			opts.BeforeEpochMs = before.UnixMilli()
		}

		returnedRows, err := spClient.PlayerRecentlyPlayedOpt(ctx, &opts)
		if err != nil {
			fmt.Println(err)
			return
		}

		if len(returnedRows) == 0 {
			break
		}

		rows = &returnedRows
		before = &returnedRows[len(returnedRows)-1].PlayedAt

		insertParams, err := processHistory(ctx, spClient, returnedRows, userID)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = history.InsertEntries(ctx, db.Service().DB, insertParams)
		if err != nil {
			fmt.Println(err)
			return
		}
		time.Sleep(5 * time.Second)
	}

}

func processHistory(ctx context.Context, spClient *z_spotify.Client, items []z_spotify.RecentlyPlayedItem, userID uuid.UUID) ([]db.HistoryInsertOneParams, error) {
	allRows := []db.HistoryInsertOneParams{}

	for i := range items {
		currentIndex := len(items) - i - 1
		item := items[currentIndex]

		durationMS := int32(item.Track.Duration)
		if currentIndex != 0 {
			nextItem := items[currentIndex-1]
			// fmt.Printf("this: %s, next: %s (%s)\n", item.PlayedAt.Format(time.DateTime), nextItem.PlayedAt.Format(time.DateTime), nextItem.Track.Name)
			startDiffMS := int32(nextItem.PlayedAt.Sub(item.PlayedAt).Milliseconds())
			if startDiffMS < durationMS {
				// fmt.Printf("%s was played for %ds / %ds (next)\n", item.Track.Name, durationMS/1000, item.Track.Duration/1000)
				durationMS = startDiffMS
			}
		} else if nowDiffMS := int32(time.Since(item.PlayedAt).Milliseconds()); nowDiffMS < durationMS {
			// fmt.Printf("%s was played for %ds / %ds (now)\n", item.Track.Name, durationMS/1000, item.Track.Duration/1000)
			durationMS = nowDiffMS
		}

		trackData, err := spotify.GetTrack(ctx, spClient, item.Track.ID.String())
		if err != nil {
			fmt.Println(err)
			continue
		}

		row := db.HistoryInsertOneParams{
			UserID:           userID,
			Timestamp:        item.PlayedAt,
			MsPlayed:         durationMS,
			TrackName:        trackData.Name,
			ArtistName:       trackData.ArtistName,
			AlbumName:        trackData.AlbumName,
			SpotifyTrackUri:  string(item.Track.URI),
			SpotifyArtistUri: sql.NullString{Valid: true, String: string(trackData.ArtistURI)},
			SpotifyAlbumUri:  sql.NullString{Valid: true, String: string(trackData.AlbumURI)},
		}

		allRows = append(allRows, row)
	}
	return allRows, nil
}
