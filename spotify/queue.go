package spotify

import (
	"context"
	"fmt"
	"time"

	"github.com/zmb3/spotify/v2"
)

type TrackInfo struct {
	ID                       string         `json:"id"`
	Name                     string         `json:"name"`
	Artists                  []string       `json:"artists"`
	Image                    *spotify.Image `json:"image"`
	AddedBy                  string         `json:"added_by,omitempty"`
	AddedAt                  time.Time      `json:"-"`
	Paused                   bool           `json:"paused"`
	DurationMS               int            `json:"duration_ms,omitempty"`
	StartedPlayingEpochMilis *int64         `json:"started_playing_epoch_ms,omitempty"`
}

type CurrentQueue struct {
	CurrentlyPlaying TrackInfo   `json:"currently_playing"`
	Queue            []TrackInfo `json:"queue"`
}

func UpdateUserPlayback(ctx context.Context, client *spotify.Client, playbackStatus *CurrentQueue) error {
	playback, err := client.PlayerState(ctx)
	if err != nil {
		return fmt.Errorf("could not get playback: %w", err)
	}
	if playback == nil || playbackStatus.CurrentlyPlaying.ID == "" {
		return nil
	}
	startedPlaying := time.Now().UnixMilli() - int64(playback.Progress)
	playbackStatus.CurrentlyPlaying.StartedPlayingEpochMilis = &startedPlaying
	playbackStatus.CurrentlyPlaying.Paused = !playback.Playing
	return nil
}

func GetUserQueue(ctx context.Context, client *spotify.Client) (*CurrentQueue, error) {
	queue, err := client.GetQueue(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get queue: %w", err)
	}
	cq := &CurrentQueue{
		CurrentlyPlaying: TrackInfo{
			ID:         queue.CurrentlyPlaying.ID.String(),
			Name:       queue.CurrentlyPlaying.Name,
			Artists:    GetArtists(queue.CurrentlyPlaying),
			Image:      Get64Image(queue.CurrentlyPlaying.Album),
			DurationMS: queue.CurrentlyPlaying.Duration,
		},
	}
	for _, entry := range queue.Items {
		qe := TrackInfo{
			ID:         entry.ID.String(),
			Name:       entry.Name,
			Artists:    GetArtists(entry),
			Image:      Get64Image(entry.Album),
			DurationMS: entry.Duration,
		}
		cq.Queue = append(cq.Queue, qe)
	}

	return cq, nil
}

func PushToUserQueue(ctx context.Context, client *spotify.Client, songID string) error {
	err := client.QueueSong(ctx, spotify.ID(songID))
	if err != nil {
		return fmt.Errorf("push to queue: %w", err)
	}
	return nil
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
