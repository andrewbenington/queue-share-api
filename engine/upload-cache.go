package engine

import (
	"context"
	"log"
	"time"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/samber/lo"
	z_spotify "github.com/zmb3/spotify/v2"
	"golang.org/x/exp/maps"
)

// func uploadTrackCache(ctx context.Context) error {
// 	cache := spotify.GetTrackCache()

// 	trackPtrs := lo.Map(
// 		maps.Values(cache),
// 		func(track z_spotify.FullTrack, _ int) *z_spotify.FullTrack { return &track },
// 	)

// 	log.Printf("There are %d tracks in the cache", len(trackPtrs))

// 	transaction, err := db.Service().DB.BeginTx(ctx, nil)
// 	if err != nil {
// 		return err
// 	}

// 	for i := range (len(trackPtrs) / 100) + 1 {
// 		start := i * 100
// 		end := start + 100
// 		if end > len(trackPtrs) {
// 			end = len(trackPtrs)
// 		}

// 		params := spotify.InsertParamsFromFullTracks(trackPtrs[start:end])

// 		err = db.New(transaction).TrackCacheInsertBulkNullable(ctx, params)
// 		if err != nil {
// 			log.Println(err)
// 		}

// 		log.Printf("Uploaded %d/%d", end, len(trackPtrs))
// 		time.Sleep(time.Millisecond)
// 	}

// 	err = transaction.Commit()
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func uploadAlbumCache(ctx context.Context) error {
	cache := spotify.GetAlbumCache()

	albumPtrs := lo.Map(
		maps.Values(cache),
		func(album z_spotify.FullAlbum, _ int) *z_spotify.FullAlbum { return &album },
	)

	log.Printf("There are %d albums in the cache", len(albumPtrs))

	transaction, err := db.Service().DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for i := range (len(albumPtrs) / 100) + 1 {
		start := i * 100
		end := start + 100
		if end > len(albumPtrs) {
			end = len(albumPtrs)
		}

		params := spotify.InsertParamsFromFullAlbums(albumPtrs[start:end])

		err = db.New(transaction).AlbumCacheInsertBulkNullable(ctx, params)
		if err != nil {
			log.Println(err)
		}

		log.Printf("Uploaded %d/%d", end, len(albumPtrs))
		time.Sleep(time.Millisecond)
	}

	err = transaction.Commit()
	if err != nil {
		return err
	}

	return nil
}

func uploadArtistCache(ctx context.Context) error {
	cache := spotify.GetArtistCache()

	artistPtrs := lo.Map(
		maps.Values(cache),
		func(artist z_spotify.FullArtist, _ int) *z_spotify.FullArtist { return &artist },
	)

	log.Printf("There are %d artists in the cache", len(artistPtrs))

	transaction, err := db.Service().DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for i := range (len(artistPtrs) / 100) + 1 {
		start := i * 100
		end := start + 100
		if end > len(artistPtrs) {
			end = len(artistPtrs)
		}

		params := spotify.InsertParamsFromFullArtists(artistPtrs[start:end])

		err = db.New(transaction).ArtistCacheInsertBulkNullable(ctx, params)
		if err != nil {
			log.Println(err)
		}

		log.Printf("Uploaded %d/%d", end, len(artistPtrs))
		time.Sleep(time.Millisecond)
	}

	err = transaction.Commit()
	if err != nil {
		return err
	}

	return nil
}
