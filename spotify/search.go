package spotify

import (
	"context"
	"fmt"

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
			Artists: GetArtists(entry),
			Image:   Get64Image(entry.Album),
		}
		resultsTrunced = append(resultsTrunced, t)
	}
	return resultsTrunced, nil
}

func GetTracks(ctx context.Context, spClient *spotify.Client, ids []string) (map[string]spotify.FullTrack, error) {

	tracks := getTracksFromCache(ids)
	idsToGet := []spotify.ID{}
	fmt.Printf("%d/%d tracks already cached", len(tracks), len(ids))

	for _, id := range ids {
		if _, ok := tracks[id]; !ok {
			idsToGet = append(idsToGet, spotify.ID(id))
		}
	}
	if len(idsToGet) > 0 {
		fmt.Printf("Getting tracks: %v\n", idsToGet)
		trackResults, err := spClient.GetTracks(ctx, idsToGet)
		fmt.Println("Got tracks")
		if err != nil {
			return nil, err
		}

		cacheTracks(trackResults)

		for _, track := range trackResults {
			tracks[track.ID.String()] = *track
		}
	}

	return tracks, nil
}
