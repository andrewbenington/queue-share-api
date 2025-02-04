package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
)

func CacheFullTracks(ctx context.Context, tx db.DBTX, tracks []*spotify.FullTrack) {
	params := InsertParamsFromFullTracks(tracks)

	err := db.New(tx).TrackCacheInsertBulkNullable(ctx, params)
	if err != nil {
		fmt.Printf("Error inserting into track cache: %s", err)
		return
	}
}

func InsertParamsFromFullTracks(tracks []*spotify.FullTrack) db.TrackCacheInsertBulkNullableParams {
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
		image := GetAlbum300Image(track.Album)
		if image != nil {
			params.ImageUrl = append(params.ImageUrl, &image.URL)
		} else {
			params.ImageUrl = append(params.ImageUrl, nil)
		}

		if len(track.Artists) > 1 {
			trackArtists := lo.Map(track.Artists[1:], trackArtistFromSimple)
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

func trackArtistFromSimple(artist spotify.SimpleArtist, _ int) db.TrackArtist {
	id := artist.ID.String()
	uri := string(artist.URI)
	return db.TrackArtist{
		Name: artist.Name,
		ID:   &id,
		URI:  &uri,
	}
}

func GetTracksFromCache(ctx context.Context, tx db.DBTX, ids []string) (map[string]db.TrackData, error) {
	rows, err := db.New(tx).TrackCacheGetByID(ctx, ids)
	if err != nil {
		return nil, err
	}
	cacheHits := map[string]db.TrackData{}
	for _, row := range rows {
		cacheHits[row.ID] = *row
	}
	return cacheHits, nil
}

func stringSlicesDiffer(slice1 []string, slice2 []string) bool {
	inBothMap := map[string]bool{}

	for _, s := range slice1 {
		inBothMap[s] = false
	}
	for _, s := range slice2 {
		if _, ok := inBothMap[s]; !ok {
			return true
		}
		inBothMap[s] = true
	}

	for _, inBoth := range inBothMap {
		if !inBoth {
			return true
		}
	}
	return false
}

func CacheSpotifyArtists(ctx context.Context, artistsToCache []*spotify.FullArtist) error {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	artistIDs := lo.Map(artistsToCache, func(a *spotify.FullArtist, _ int) string { return a.ID.String() })

	rows, err := db.New(tx).ArtistCacheGetByID(ctx, artistIDs)
	if err != nil {
		return err
	}

	dbArtists := lo.SliceToMap(rows, func(row *db.ArtistData) (string, *db.ArtistData) {
		return row.ID, row
	})

	for i, artistToCache := range artistsToCache {
		if i%100 == 0 {
			log.Printf("processing artist %d/%d", i+1, len(artistsToCache))
		}
		var preferredImageURL *string
		if preferredImage := GetArtist128Image(*artistToCache); preferredImage != nil {
			preferredImageURL = &preferredImage.URL
		}

		popularity := int32(artistToCache.Popularity)
		followerCount := int32(artistToCache.Followers.Count)

		if dbArtist, ok := dbArtists[artistToCache.ID.String()]; ok {
			genresHaveChanges := stringSlicesDiffer(dbArtist.Genres, artistToCache.Genres)
			var dbPopularity int32 = 0
			if dbArtist.Popularity != nil {
				dbPopularity = *dbArtist.Popularity
			}
			var dbFollowers int32 = 0
			if dbArtist.FollowerCount != nil {
				dbPopularity = *dbArtist.FollowerCount
			}
			if dbArtist.ImageUrl != preferredImageURL ||
				dbArtist.Name != artistToCache.Name ||
				genresHaveChanges ||
				dbPopularity != popularity ||
				dbFollowers != followerCount {
				log.Printf("updating artist cache for %s", dbArtist.Name)

				db.New(tx).ArtistCacheUpdateOne(ctx, db.ArtistCacheUpdateOneParams{
					ID:            artistToCache.ID.String(),
					Name:          artistToCache.Name,
					ImageUrl:      preferredImageURL,
					Genres:        artistToCache.Genres,
					Popularity:    &popularity,
					FollowerCount: &followerCount,
				})
			}
		} else {
			log.Printf("inserting new entry for %s", artistToCache.Name)
			db.New(tx).ArtistCacheInsertOne(ctx, db.ArtistCacheInsertOneParams{
				ID:            artistToCache.ID.String(),
				URI:           string(artistToCache.URI),
				Name:          artistToCache.Name,
				ImageUrl:      preferredImageURL,
				Genres:        artistToCache.Genres,
				Popularity:    &popularity,
				FollowerCount: &followerCount,
			})
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func getArtistsFromCache(ctx context.Context, tx db.DBTX, ids []string) (map[string]db.ArtistData, error) {
	rows, err := db.New(tx).ArtistCacheGetByID(ctx, ids)
	if err != nil {
		return nil, err
	}
	cacheHits := map[string]db.ArtistData{}
	for _, row := range rows {
		cacheHits[row.ID] = *row
	}
	return cacheHits, nil
}

func strPtrsDiffer(p1 *string, p2 *string) bool {
	if p1 == nil && p2 == nil {
		return false
	}
	if p1 == nil || p2 == nil {
		return true
	}
	return *p1 == *p2
}

func i32PtrsDiffer(p1 *int32, p2 *int32) bool {
	if p1 == nil && p2 == nil {
		return false
	}
	if p1 == nil || p2 == nil {
		return true
	}
	return *p1 == *p2
}

func CacheSpotifyAlbums(ctx context.Context, albumsToCache []*spotify.FullAlbum) error {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	albumIDs := lo.Map(albumsToCache, func(a *spotify.FullAlbum, _ int) string { return a.ID.String() })

	rows, err := db.New(tx).AlbumCacheGetByID(ctx, albumIDs)
	if err != nil {
		return err
	}

	dbAlbums := lo.SliceToMap(rows, func(row *db.AlbumData) (string, *db.AlbumData) {
		return row.ID, row
	})

	for i, spotifyAlbum := range albumsToCache {
		if i%100 == 0 {
			log.Printf("processing album %d/%d", i+1, len(albumsToCache))
		}

		newData, err := AlbumDataFromSpotifyAlbum(*spotifyAlbum)
		if err != nil {
			log.Printf("Error converting spotify album %s (%s): %s", spotifyAlbum.Name, spotifyAlbum.URI, err)
			continue
		}

		if existingData, ok := dbAlbums[spotifyAlbum.ID.String()]; ok {
			genresHaveChanges := stringSlicesDiffer(existingData.Genres, newData.Genres)

			trackIDsHaveChanges := stringSlicesDiffer(existingData.SpotifyTrackIds, newData.SpotifyTrackIds)
			trackISRCsHaveChanges := stringSlicesDiffer(existingData.TrackIsrcs, newData.TrackIsrcs)

			if strPtrsDiffer(existingData.ImageUrl, newData.ImageUrl) ||
				existingData.Name != newData.Name ||
				genresHaveChanges ||
				trackIDsHaveChanges || trackISRCsHaveChanges ||
				i32PtrsDiffer(existingData.Popularity, newData.Popularity) {
				log.Printf("updating album cache for %s", existingData.Name)

				db.New(tx).AlbumCacheUpdateOne(ctx, db.AlbumCacheUpdateOneParams{
					ID:              newData.ID,
					Name:            newData.Name,
					ImageUrl:        newData.ImageUrl,
					Genres:          newData.Genres,
					Popularity:      newData.Popularity,
					Upc:             newData.Upc,
					SpotifyTrackIds: newData.SpotifyTrackIds,
					TrackIsrcs:      newData.TrackIsrcs,
				})
			}
		} else {
			log.Printf("inserting new entry for %s", spotifyAlbum.Name)

			db.New(tx).AlbumCacheInsertOne(ctx, db.AlbumCacheInsertOneParams{
				ID:                   newData.ID,
				URI:                  newData.URI,
				Name:                 newData.Name,
				ArtistID:             newData.ArtistID,
				ArtistURI:            newData.ArtistURI,
				ArtistName:           newData.ArtistName,
				AlbumGroup:           newData.AlbumGroup,
				AlbumType:            newData.AlbumType,
				ImageUrl:             newData.ImageUrl,
				ReleaseDate:          newData.ReleaseDate,
				ReleaseDatePrecision: newData.ReleaseDatePrecision,
				Genres:               newData.Genres,
				Popularity:           newData.Popularity,
				Upc:                  newData.Upc,
				SpotifyTrackIds:      newData.SpotifyTrackIds,
				TrackIsrcs:           newData.TrackIsrcs,
			})
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func InsertParamsFromFullAlbums(albums []*spotify.FullAlbum) db.AlbumCacheInsertBulkNullableParams {
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
		image := GetAlbum300Image(album.SimpleAlbum)
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

func getAlbumsFromCache(ctx context.Context, tx db.DBTX, ids []string) (map[string]db.AlbumData, error) {
	rows, err := db.New(tx).AlbumCacheGetByID(ctx, ids)
	if err != nil {
		return nil, err
	}
	cacheHits := map[string]db.AlbumData{}
	for _, row := range rows {
		cacheHits[row.ID] = *row
	}
	return cacheHits, nil
}
