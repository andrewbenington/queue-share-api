package spotify

import (
	"context"
	"fmt"
	"log"

	"github.com/andrewbenington/queue-share-api/db"
	"github.com/zmb3/spotify/v2"
)

func SearchSongs(ctx context.Context, spClient *spotify.Client, text string) ([]TrackInfo, error) {
	results, err := spClient.Search(ctx, text, spotify.SearchTypeTrack)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	resultsTrunced := []TrackInfo{}
	for _, entry := range results.Tracks.Tracks {
		t := TrackInfo{
			ID:      entry.ID.String(),
			Name:    entry.Name,
			Artists: entry.Artists,
			Image:   GetAlbum64Image(entry.Album),
		}
		resultsTrunced = append(resultsTrunced, t)
	}
	return resultsTrunced, nil
}

func GetTrack(ctx context.Context, spClient *spotify.Client, id string) (*db.TrackData, error) {

	tracks, err := getTracksFromCache(ctx, []string{id})
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

	cacheFullTracks(ctx, []*spotify.FullTrack{track})

	trackData := TrackDataFromFullTrack(*track)

	return &trackData, nil
}

func GetTracks(ctx context.Context, spClient *spotify.Client, ids []string) (map[string]db.TrackData, error) {
	tracks, err := getTracksFromCache(ctx, ids)
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

		cacheFullTracks(ctx, results)

		for _, track := range results {
			tracks[track.ID.String()] = TrackDataFromFullTrack(*track)
		}
	}

	return tracks, nil
}

func GetArtist(ctx context.Context, spClient *spotify.Client, id string) (*spotify.FullArtist, error) {

	artists := getArtistsFromCache([]string{id})

	if artist, ok := artists[id]; ok {
		return &artist, nil
	}

	artist, err := spClient.GetArtist(ctx, spotify.ID(id))
	if err != nil {
		return nil, err
	}

	cacheArtists([]*spotify.FullArtist{artist})

	return artist, nil
}

func GetArtists(ctx context.Context, spClient *spotify.Client, ids []string) (map[string]spotify.FullArtist, error) {

	artists := getArtistsFromCache(ids)
	idsToGet := []spotify.ID{}
	fmt.Printf("%d/%d artists already cached\n", len(artists), len(ids))

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

		cacheArtists(results)

		for _, artist := range results {
			artists[artist.ID.String()] = *artist
		}
	}

	return artists, nil
}

func GetAlbum(ctx context.Context, spClient *spotify.Client, id string) (*spotify.FullAlbum, error) {

	albums := getAlbumsFromCache([]string{id})

	if album, ok := albums[id]; ok {
		return &album, nil
	}

	album, err := spClient.GetAlbum(ctx, spotify.ID(id))
	if err != nil {
		return nil, err
	}

	cacheAlbums(ctx, []*spotify.FullAlbum{album})

	return album, nil
}

func GetAlbums(ctx context.Context, spClient *spotify.Client, ids []string) (map[string]spotify.FullAlbum, error) {
	albums := getAlbumsFromCache(ids)
	idsToGet := []spotify.ID{}
	fmt.Printf("%d/%d albums already cached\n", len(albums), len(ids))

	for _, id := range ids {
		if _, ok := albums[id]; !ok {
			idsToGet = append(idsToGet, spotify.ID(id))
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
		for _, result := range results {
			result.Tracks = spotify.SimpleTrackPage{}
		}

		cacheAlbums(ctx, results)

		for _, album := range results {
			albums[album.ID.String()] = *album
		}
	}

	return albums, nil
}
