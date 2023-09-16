package queue

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

func PushToUserQueue(ctx context.Context, client *spotify.Client, songID string) error {
	err := client.QueueSong(ctx, spotify.ID(songID))
	if err != nil {
		return fmt.Errorf("push to queue: %w", err)
	}
	return nil
}
