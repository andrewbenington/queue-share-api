package queue

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

func PushToUserQueue(client *spotify.Client, songID string) error {
	err := client.QueueSong(context.Background(), spotify.ID(songID))
	if err != nil {
		return fmt.Errorf("push to queue: %w", err)
	}
	return nil
}
