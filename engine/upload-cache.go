package engine

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

// func uploadAlbumCache(ctx context.Context) error {
// 	cache := service.GetAlbumCache()

// 	albumPtrs := lo.Map(
// 		maps.Values(cache),
// 		func(album spotify.FullAlbum, _ int) *spotify.FullAlbum { return &album },
// 	)

// 	log.Printf("There are %d albums in the cache", len(albumPtrs))

// 	return service.CacheSpotifyAlbums(ctx, albumPtrs)
// }
