package engine

import (
	"context"
	"log"
	"time"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/exp/maps"
)

// func uploadTrackCache(ctx context.Context) error {
// 	cache := service.GetTrackCache()

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

// 		params := service.InsertParamsFromFullTracks(trackPtrs[start:end])

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
	cache := service.GetAlbumCache()

	albumPtrs := lo.Map(
		maps.Values(cache),
		func(album spotify.FullAlbum, _ int) *spotify.FullAlbum { return &album },
	)

	log.Printf("There are %d albums in the cache", len(albumPtrs))

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i := range (len(albumPtrs) / 100) + 1 {
		start := i * 100
		end := start + 100
		if end > len(albumPtrs) {
			end = len(albumPtrs)
		}

		params := service.InsertParamsFromFullAlbums(albumPtrs[start:end])

		err = db.New(tx).AlbumCacheInsertBulkNullable(ctx, params)
		if err != nil {
			log.Println(err)
		}

		log.Printf("Uploaded %d/%d", end, len(albumPtrs))
		time.Sleep(time.Millisecond)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func uploadArtistCache(ctx context.Context) error {
	cache := service.GetArtistCache()

	artistPtrs := lo.Map(
		maps.Values(cache),
		func(artist spotify.FullArtist, _ int) *spotify.FullArtist { return &artist },
	)

	log.Printf("There are %d artists in the cache", len(artistPtrs))

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i := range (len(artistPtrs) / 100) + 1 {
		start := i * 100
		end := start + 100
		if end > len(artistPtrs) {
			end = len(artistPtrs)
		}

		params := service.InsertParamsFromFullArtists(artistPtrs[start:end])

		err = db.New(tx).ArtistCacheInsertBulkNullable(ctx, params)
		if err != nil {
			log.Println(err)
		}

		log.Printf("Uploaded %d/%d", end, len(artistPtrs))
		time.Sleep(time.Millisecond)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
