package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
)

var (
	spotifyAlbumCache = map[string]spotify.FullAlbum{}
)

func init() {

	// load album cache
	filePath := path.Join("temp", "album_cache.json")
	if _, err := os.Stat(filePath); err != nil {
		return
	}
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error loading track data cache: %s\n", err)
		return
	}
	err = json.Unmarshal(bytes, &spotifyAlbumCache)
	if err != nil {
		fmt.Printf("Error parsing album data cache: %s\n", err)
	} else {
		fmt.Println("Album cache loaded successfully")
	}
}

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

func stringSlicesAreEqual(s1 []string, s2 []string) bool {
	inFirst := map[string]bool{}
	for _, s := range s1 {
		inFirst[s] = false
	}

	for _, s := range s2 {
		if _, ok := inFirst[s]; ok {
			inFirst[s] = true
		} else {
			return false
		}
	}

	for _, inSecond := range inFirst {
		if !inSecond {
			return false
		}
	}
	return true
}

func CacheArtistsPtrs(ctx context.Context, artistsToCache []*spotify.FullArtist) error {
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
			genresHaveChanges := stringSlicesAreEqual(dbArtist.Genres, artistToCache.Genres)
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

func CacheArtists(ctx context.Context, artistsToCache []spotify.FullArtist) error {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	artistIDs := lo.Map(artistsToCache, func(a spotify.FullArtist, _ int) string { return a.ID.String() })

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
		if preferredImage := GetArtist128Image(artistToCache); preferredImage != nil {
			preferredImageURL = &preferredImage.URL
		}

		popularity := int32(artistToCache.Popularity)
		followerCount := int32(artistToCache.Followers.Count)

		if dbArtist, ok := dbArtists[artistToCache.ID.String()]; ok {
			genresHaveChanges := stringSlicesAreEqual(dbArtist.Genres, artistToCache.Genres)
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

func cacheFullAlbums(ctx context.Context, tx db.DBTX, albums []*spotify.FullAlbum) {
	for _, album := range albums {
		spotifyAlbumCache[string(album.ID)] = *album
	}
	bytes, err := json.Marshal(spotifyAlbumCache)
	if err != nil {
		fmt.Printf("Error serializing Spotify album cache: %s", err)
		return
	}
	err = os.WriteFile(path.Join("temp", "album_cache.json"), bytes, 0644)
	if err != nil {
		fmt.Printf("Error serializing Spotify album cache: %s", err)
	}
	params := InsertParamsFromFullAlbums(albums)
	err = db.New(tx).AlbumCacheInsertBulkNullable(ctx, params)
	if err != nil {
		fmt.Printf("Error inserting into album cache: %s", err)
		return
	}
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

func getAlbumsFromCache(ids []string) map[string]spotify.FullAlbum {
	cacheHits := map[string]spotify.FullAlbum{}
	for _, id := range ids {
		if albumData, ok := spotifyAlbumCache[id]; ok {
			cacheHits[id] = albumData
		}
	}
	return cacheHits
}

func GetAlbumCache() map[string]spotify.FullAlbum {
	return spotifyAlbumCache
}

func TrackDataFromFullTrackIdx(ft spotify.FullTrack, _ int) db.TrackData {
	return TrackDataFromFullTrack(ft)
}

func TrackDataFromFullTrack(ft spotify.FullTrack) db.TrackData {
	artist := ft.Artists[0]

	image := GetAlbum300Image(ft.Album)
	var imageURL *string
	if image != nil {
		imageURL = &image.URL
	}

	var isrc *string
	if isrcField, ok := ft.ExternalIDs["isrc"]; ok {
		isrc = &isrcField
	}

	return db.TrackData{
		ID:           ft.ID.String(),
		URI:          string(ft.URI),
		Name:         ft.Name,
		AlbumID:      ft.Album.ID.String(),
		AlbumURI:     string(ft.Album.URI),
		AlbumName:    ft.Album.Name,
		ArtistID:     artist.ID.String(),
		ArtistURI:    string(artist.URI),
		ArtistName:   artist.Name,
		ImageUrl:     imageURL,
		OtherArtists: lo.Map(ft.Artists[1:], trackArtistFromSimple),
		DurationMs:   int32(ft.Duration),
		Popularity:   int32(ft.Popularity),
		Explicit:     ft.Explicit,
		PreviewUrl:   ft.PreviewURL,
		DiscNumber:   int32(ft.DiscNumber),
		TrackNumber:  int32(ft.TrackNumber),
		Type:         ft.Type,
		ExternalIds:  ft.ExternalIDs,
		Isrc:         isrc,
	}
}

func ArtistDataFromFullArtistIdx(fa spotify.FullArtist, _ int) db.ArtistData {
	return ArtistDataFromFullArtist(fa)
}

func ArtistDataFromFullArtist(fa spotify.FullArtist) db.ArtistData {
	image := GetArtist300Image(fa)
	var imageURL *string
	if image != nil {
		imageURL = &image.URL
	}

	popularity := int32(fa.Popularity)
	followerCount := int32(fa.Followers.Count)

	return db.ArtistData{
		ID:            fa.ID.String(),
		URI:           string(fa.URI),
		Name:          fa.Name,
		ImageUrl:      imageURL,
		Genres:        fa.Genres,
		Popularity:    &popularity,
		FollowerCount: &followerCount,
	}
}
