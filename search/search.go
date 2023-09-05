package search

import (
	"context"
	"fmt"

	"github.com/andrewbenington/go-spotify/queue"
	"github.com/zmb3/spotify/v2"
)

func SearchSongs(client *spotify.Client, text string) ([]queue.TrackInfo, error) {
	results, err := client.Search(context.Background(), text, spotify.SearchTypeTrack)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}
	resultsTrunced := []queue.TrackInfo{}
	for _, entry := range results.Tracks.Tracks {
		t := queue.TrackInfo{
			ID:      entry.ID.String(),
			Name:    entry.Name,
			Artists: queue.GetArtists(entry),
			Image:   queue.Get64Image(entry.Album),
		}
		resultsTrunced = append(resultsTrunced, t)
	}
	return resultsTrunced, nil
}
