package spotify

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

type SpotifyUser struct {
	ID       string
	Display  string
	ImageURL string
}

func GetUser(ctx context.Context, client *spotify.Client) (*SpotifyUser, error) {
	user, err := client.CurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get user: %w", err)
	}
	var userImage spotify.Image
	for i := range user.Images {
		if userImage.Height == 0 || user.Images[i].Height < userImage.Height {
			userImage = user.Images[i]
		}
	}
	u := &SpotifyUser{
		ID:       user.ID,
		ImageURL: userImage.URL,
		Display:  user.DisplayName,
	}

	return u, nil
}
