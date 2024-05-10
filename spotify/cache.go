package spotify

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
	spotifyTrackCache  = map[string]spotify.FullTrack{}
	spotifyArtistCache = map[string]spotify.FullArtist{}
	spotifyAlbumCache  = map[string]spotify.FullAlbum{}
)

func init() {
	// load track cache
	filePath := path.Join("temp", "track_cache.json")
	if _, err := os.Stat(filePath); err != nil {
		return
	}
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error loading track data cache: %s\n", err)
		return
	}
	err = json.Unmarshal(bytes, &spotifyTrackCache)
	if err != nil {
		fmt.Printf("Error parsing track data cache: %s\n", err)
	} else {
		fmt.Println("Track cache loaded successfully")
	}

	// load artist cache
	filePath = path.Join("temp", "artist_cache.json")
	if _, err := os.Stat(filePath); err != nil {
		return
	}
	bytes, err = os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error loading track data cache: %s\n", err)
		return
	}
	err = json.Unmarshal(bytes, &spotifyArtistCache)
	if err != nil {
		fmt.Printf("Error parsing artist data cache: %s\n", err)
	} else {
		fmt.Println("Artist cache loaded successfully")
	}

	// load album cache
	filePath = path.Join("temp", "album_cache.json")
	if _, err := os.Stat(filePath); err != nil {
		return
	}
	bytes, err = os.ReadFile(filePath)
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

func CacheFullTracks(ctx context.Context, tracks []*spotify.FullTrack) {
	params := InsertParamsFromFullTracks(tracks)
	err := db.New(db.Service().DB).TrackCacheInsertBulkNullable(ctx, params)
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
	return db.TrackArtist{
		Name: artist.Name,
		ID:   artist.ID.String(),
		URI:  string(artist.URI),
	}
}

func getTracksFromCache(ctx context.Context, ids []string) (map[string]db.TrackData, error) {
	rows, err := db.New(db.Service().DB).TrackCacheGetByID(ctx, ids)
	if err != nil {
		return nil, err
	}
	cacheHits := map[string]db.TrackData{}
	for _, row := range rows {
		cacheHits[row.ID] = *row
	}
	return cacheHits, nil
}

func cacheArtists(artists []*spotify.FullArtist) {
	for _, artist := range artists {
		spotifyArtistCache[string(artist.ID)] = *artist
	}
	bytes, err := json.Marshal(spotifyArtistCache)
	if err != nil {
		fmt.Printf("Error serializing Spotify artist cache: %s", err)
		return
	}
	err = os.WriteFile(path.Join("temp", "artist_cache.json"), bytes, 0644)
	if err != nil {
		fmt.Printf("Error serializing Spotify artist cache: %s", err)
	}
}

func InsertParamsFromFullArtists(artists []*spotify.FullArtist) db.ArtistCacheInsertBulkNullableParams {
	params := db.ArtistCacheInsertBulkNullableParams{
		ID:            []string{},
		Uri:           []string{},
		Name:          []string{},
		ImageUrl:      []*string{},
		Genres:        []*string{},
		Popularity:    []int32{},
		FollowerCount: []int32{},
	}

	for _, artist := range artists {
		params.ID = append(params.ID, artist.ID.String())
		params.Uri = append(params.Uri, string(artist.URI))
		params.Name = append(params.Name, artist.Name)
		image := GetArtist300Image(*artist)
		if image != nil {
			params.ImageUrl = append(params.ImageUrl, &image.URL)
		} else {
			params.ImageUrl = append(params.ImageUrl, nil)
		}

		bytes, err := json.Marshal(artist.Genres)
		if err != nil {
			log.Println(err)
			params.Genres = append(params.Genres, nil)
		} else {
			str := string(bytes)
			params.Genres = append(params.Genres, &str)
		}

		params.Popularity = append(params.Popularity, int32(artist.Popularity))
		params.FollowerCount = append(params.FollowerCount, int32(artist.Followers.Count))
	}

	return params
}

func getArtistsFromCache(ids []string) map[string]spotify.FullArtist {
	cacheHits := map[string]spotify.FullArtist{}
	for _, id := range ids {
		if artistData, ok := spotifyArtistCache[id]; ok {
			cacheHits[id] = artistData
		}
	}
	return cacheHits
}

func cacheAlbums(ctx context.Context, albums []*spotify.FullAlbum) {
	params := InsertParamsFromFullAlbums(albums)
	err := db.New(db.Service().DB).AlbumCacheInsertBulkNullable(ctx, params)
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

func GetTrackCache() map[string]spotify.FullTrack {
	return spotifyTrackCache
}

func GetAlbumCache() map[string]spotify.FullAlbum {
	return spotifyAlbumCache
}

func GetArtistCache() map[string]spotify.FullArtist {
	return spotifyArtistCache
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
