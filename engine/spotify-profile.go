package engine

import (
	"context"
	"fmt"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
)

func doSpotifyProfileCycle(ctx context.Context) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		fmt.Printf("Could not connect to database to get missing artist URIs %s\n", err)
		return
	}
	defer tx.Commit(ctx)

	users, err := db.New(tx).UserGetAllWithSpotify(ctx)
	if err != nil {
		fmt.Printf("Could not get users with Spotify profiles: %s\n", err)
		return
	}
	for _, user := range users {
		_, spClient, err := client.ForUser(ctx, user.ID)
		if err != nil {
			fmt.Printf("Could not get client for user %s: %s\n", user.ID, err)
			continue
		}

		userData, err := spClient.CurrentUser(ctx)
		if err != nil {
			fmt.Printf("Could not get spotify data for user %s: %s\n", user.ID, err)
			continue
		}

		if (user.SpotifyAccount != nil && *user.SpotifyAccount != userData.ID) ||
			(user.SpotifyName != nil && *user.SpotifyName != userData.User.DisplayName) ||
			(len(userData.Images) > 0 && user.SpotifyImageUrl != nil && *user.SpotifyImageUrl != userData.Images[0].URL) {
			err = db.New(tx).UserUpdateSpotifyInfo(ctx, db.UserUpdateSpotifyInfoParams{
				ID:              user.ID,
				SpotifyAccount:  &userData.ID,
				SpotifyName:     &userData.User.DisplayName,
				SpotifyImageUrl: &userData.Images[0].URL,
			})
			if err != nil {
				fmt.Printf("Could not update spotify data for user %s: %s\n", user.ID, err)
			}
		}
	}
}
