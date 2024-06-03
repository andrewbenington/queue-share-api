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
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
)

var (
	LastFetch *time.Time
)

func doHistoryCycle(ctx context.Context) {
	users, err := db.New(db.Service().DB).UsersToFetchHistory(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, user := range users {
		cancelCtx, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()
		getHistoryForUser(cancelCtx, user)
		log.Printf("History fetched for user %s\n", user.Username)
	}
	now := time.Now()
	LastFetch = &now
}

func getHistoryForUser(ctx context.Context, user *db.User) {
	_, spClient, err := client.ForUser(ctx, user.ID)
	if err != nil {
		fmt.Println(err)
		return
	}

	var rows *[]spotify.RecentlyPlayedItem
	var before *time.Time

	for rows == nil || len(*rows) > 0 {
		opts := spotify.RecentlyPlayedOptions{Limit: 50}
		if before != nil {
			opts.BeforeEpochMs = before.UnixMilli()
		}

		recentlyPlayed, err := spClient.PlayerRecentlyPlayedOpt(ctx, &opts)
		if err != nil {
			fmt.Println(err)
			return
		} else {
			log.Printf("%d history entries found for %s", len(recentlyPlayed), user.Username)
		}

		// Load URIs
		trackIDs := lo.Map(recentlyPlayed, func(track spotify.RecentlyPlayedItem, _ int) string {
			return track.Track.ID.String()
		})
		service.GetTracks(ctx, spClient, lo.Uniq(trackIDs))

		albumIDs := lo.Map(recentlyPlayed, func(track spotify.RecentlyPlayedItem, _ int) string {
			return track.Track.Album.ID.String()
		})
		service.GetAlbums(ctx, spClient, lo.Uniq(albumIDs))

		artistIDs := lo.Map(recentlyPlayed, func(track spotify.RecentlyPlayedItem, _ int) string {
			return track.Track.Artists[0].ID.String()
		})
		service.GetArtists(ctx, spClient, lo.Uniq(artistIDs))

		if len(recentlyPlayed) == 0 {
			break
		}

		rows = &recentlyPlayed
		before = &recentlyPlayed[len(recentlyPlayed)-1].PlayedAt

		insertParams, err := processHistory(ctx, spClient, recentlyPlayed, user.ID)
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

func processHistory(ctx context.Context, spClient *spotify.Client, items []spotify.RecentlyPlayedItem, userID uuid.UUID) ([]db.HistoryInsertOneParams, error) {
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

		trackData, err := service.GetTrack(ctx, spClient, item.Track.ID.String())
		if err != nil {
			log.Printf("error getting track: %s", err)
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
