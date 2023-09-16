package queue

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

type TrackInfo struct {
	ID      string         `json:"id"`
	Name    string         `json:"name"`
	Artists []string       `json:"artists"`
	Image   *spotify.Image `json:"image"`
}

type CurrentQueue struct {
	CurrentlyPlaying TrackInfo   `json:"currently_playing"`
	Queue            []TrackInfo `json:"queue"`
}

func GetUserQueue(ctx context.Context, client *spotify.Client) (*CurrentQueue, error) {
	queue, err := client.GetQueue(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get queue: %w", err)
	}
	cq := &CurrentQueue{
		CurrentlyPlaying: TrackInfo{
			ID:      queue.CurrentlyPlaying.ID.String(),
			Name:    queue.CurrentlyPlaying.Name,
			Artists: GetArtists(queue.CurrentlyPlaying),
			Image:   Get64Image(queue.CurrentlyPlaying.Album),
		},
	}
	for _, entry := range queue.Items {
		qe := TrackInfo{
			ID:      entry.ID.String(),
			Name:    entry.Name,
			Artists: GetArtists(entry),
			Image:   Get64Image(entry.Album),
		}
		cq.Queue = append(cq.Queue, qe)
	}

	return cq, nil
}

func Get64Image(t spotify.SimpleAlbum) *spotify.Image {
	for i := range t.Images {
		if t.Images[i].Height == 64 {
			return &t.Images[i]
		}
	}
	return nil
}

func GetArtists(t spotify.FullTrack) []string {
	artists := []string{}
	for _, a := range t.Artists {
		artists = append(artists, a.Name)
	}
	return artists
}
