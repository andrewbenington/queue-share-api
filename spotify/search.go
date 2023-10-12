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
