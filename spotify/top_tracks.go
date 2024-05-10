package spotify

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

func TopTracks(ctx context.Context, spClient *spotify.Client) ([]TrackInfo, error) {
	results, err := spClient.CurrentUsersTopTracks(ctx)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	resultsTrunced := []TrackInfo{}
	for _, entry := range results.Tracks {
		t := TrackInfo{
			ID:      entry.ID.String(),
			Name:    entry.Name,
			Artists: entry.Artists,
			Image:   Get64Image(entry.Album),
		}
		resultsTrunced = append(resultsTrunced, t)
	}
	return resultsTrunced, nil
}
