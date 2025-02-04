package service

import (
	"context"
	"fmt"
	"log"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/util"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
)

func SearchTracks(ctx context.Context, spClient *spotify.Client, text string) ([]db.TrackData, error) {
	results, err := spClient.Search(ctx, text, spotify.SearchTypeTrack)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	return lo.Map(results.Tracks.Tracks, TrackDataFromFullTrackIdx), nil
}

func SearchArtists(ctx context.Context, spClient *spotify.Client, text string) ([]db.ArtistData, error) {
	results, err := spClient.Search(ctx, text, spotify.SearchTypeArtist)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	artistPtrs := make([]*spotify.FullArtist, 0, len(results.Artists.Artists))
	for _, a := range results.Artists.Artists {
		artistPtrs = append(artistPtrs, &a)
	}

	err = CacheSpotifyArtists(ctx, artistPtrs)
	if err != nil {
		log.Printf("error caching artists: %s", err)
	}

	return lo.Map(results.Artists.Artists, ArtistDataFromFullArtistIdx), nil
}

func SearchAlbums(ctx context.Context, spClient *spotify.Client, text string) ([]SimpleAlbumData, error) {
	results, err := spClient.Search(ctx, text, spotify.SearchTypeAlbum)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	simpleAlbums := []SimpleAlbumData{}
	for _, a := range results.Albums.Albums {
		simpleAlbum, err := SimpleAlbumDataFromSimpleAlbum(a)
		if err != nil {
			log.Printf("error converting simple album: %s", err)
			continue
		}
		simpleAlbums = append(simpleAlbums, simpleAlbum)
	}

	return simpleAlbums, nil
}

func GetTrack(ctx context.Context, spClient *spotify.Client, id string) (*db.TrackData, error) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Commit(ctx)

	tracks, err := GetTracksFromCache(ctx, tx, []string{id})
	if err != nil {
		return nil, err
	}

	if track, ok := tracks[id]; ok {
		return &track, nil
	}

	track, err := spClient.GetTrack(ctx, spotify.ID(id))
	if err != nil {
		return nil, err
	}
	fmt.Printf("Cache miss: %s\n", track.Name)

	CacheFullTracks(ctx, tx, []*spotify.FullTrack{track})

	trackData := TrackDataFromFullTrack(*track)

	return &trackData, nil
}

func GetTracks(ctx context.Context, spClient *spotify.Client, ids []string) (map[string]db.TrackData, error) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Commit(ctx)

	tracks, err := GetTracksFromCache(ctx, tx, ids)
	if err != nil {
		log.Printf("error getting tracks from cache: %s", err)
	} else {
		log.Printf("%d/%d tracks already cached", len(tracks), len(ids))
	}

	idsToGet := []spotify.ID{}

	for _, id := range ids {
		if _, ok := tracks[id]; !ok {
			idsToGet = append(idsToGet, spotify.ID(id))
		}
	}

	for start := 0; start < len(idsToGet); start += 50 {
		end := start + 50
		if end > len(idsToGet) {
			end = len(idsToGet)
		}

		results, err := spClient.GetTracks(ctx, idsToGet[start:end])
		if err != nil {
			return nil, err
		}

		if results == nil {
			continue
		}
		results = lo.Filter(results, util.PointerNotNilIdx)

		if results != nil {
			CacheFullTracks(ctx, tx, results)
		}

		for _, track := range results {
			tracks[track.ID.String()] = TrackDataFromFullTrack(*track)
		}
	}

	return tracks, nil
}

func GetArtist(ctx context.Context, spClient *spotify.Client, id string) (*db.ArtistData, error) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Commit(ctx)

	artists, err := getArtistsFromCache(ctx, tx, []string{id})
	if err != nil {
		return nil, err
	}

	if artist, ok := artists[id]; ok {
		return &artist, nil
	}

	spotifyArtist, err := spClient.GetArtist(ctx, spotify.ID(id))
	if err != nil {
		return nil, err
	}

	CacheSpotifyArtists(ctx, []*spotify.FullArtist{spotifyArtist})

	artist := ArtistDataFromFullArtist(*spotifyArtist)

	return &artist, nil
}

func GetArtists(ctx context.Context, spClient *spotify.Client, ids []string) (map[string]db.ArtistData, error) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Commit(ctx)

	artists, err := getArtistsFromCache(ctx, tx, ids)
	if err != nil {
		log.Printf("error getting artists from cache: %s", err)
	} else {
		log.Printf("%d/%d artists already cached", len(artists), len(ids))
	}

	idsToGet := []spotify.ID{}
	log.Printf("%d/%d artists already cached", len(artists), len(ids))

	for _, id := range ids {
		if _, ok := artists[id]; !ok {
			idsToGet = append(idsToGet, spotify.ID(id))
		}
	}

	for start := 0; start < len(idsToGet); start += 50 {
		end := start + 50
		if end > len(idsToGet) {
			end = len(idsToGet)
		}

		results, err := spClient.GetArtists(ctx, idsToGet[start:end]...)
		if err != nil {
			return nil, err
		}

		CacheSpotifyArtists(ctx, results)

		for _, artist := range results {
			artists[artist.ID.String()] = ArtistDataFromFullArtist(*artist)
		}
	}

	return artists, nil
}

func GetAlbum(ctx context.Context, spClient *spotify.Client, id string) (*db.AlbumData, error) {

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Commit(ctx)

	albums, err := getAlbumsFromCache(ctx, tx, []string{id})
	if err != nil {
		return nil, err
	}

	if album, ok := albums[id]; ok {
		if len(album.SpotifyTrackIds) > 0 {
			return &album, nil

		} else {
			log.Printf("Fetching %s from spotify to get track ids", album.Name)
		}
	}

	album, err := spClient.GetAlbum(ctx, spotify.ID(id))
	if err != nil {
		return nil, err
	}

	CacheSpotifyAlbums(ctx, []*spotify.FullAlbum{album})

	return AlbumDataFromSpotifyAlbum(*album)
}

func GetAlbums(ctx context.Context, spClient *spotify.Client, ids []string) (map[string]db.AlbumData, error) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Commit(ctx)

	albums, err := getAlbumsFromCache(ctx, tx, ids)
	if err != nil {
		return nil, err
	}

	idsToGet := []spotify.ID{}
	log.Printf("%d/%d albums already cached", len(albums), len(ids))

	for _, id := range ids {
		if album, ok := albums[id]; !ok {
			idsToGet = append(idsToGet, spotify.ID(id))
		} else if len(album.SpotifyTrackIds) == 0 {
			idsToGet = append(idsToGet, spotify.ID(id))
			log.Printf("Fetching %s from spotify to get track ids", album.Name)
		}
	}

	for start := 0; start < len(idsToGet); start += 20 {
		end := start + 20
		if end > len(idsToGet) {
			end = len(idsToGet)
		}

		results, err := spClient.GetAlbums(ctx, idsToGet[start:end])
		if err != nil {
			return nil, err
		}

		CacheSpotifyAlbums(ctx, results)

		for _, album := range results {
			albumData, err := AlbumDataFromSpotifyAlbum(*album)
			if err == nil {
				albums[album.ID.String()] = *albumData
			}
		}
	}

	return albums, nil
}
