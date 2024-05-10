package engine

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
)

func doSpotifyProfileCycle() {
	ctx := context.Background()
	users, err := db.New(db.Service().DB).UserGetAllWithSpotify(ctx)
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

		if user.SpotifyAccount.String != userData.ID || user.SpotifyName.String != userData.User.DisplayName || (len(userData.Images) > 0 && user.SpotifyImageUrl.String != userData.Images[0].URL) {
			err = db.New(db.Service().DB).UserUpdateSpotifyInfo(ctx, db.UserUpdateSpotifyInfoParams{
				ID:              user.ID,
				SpotifyAccount:  sql.NullString{Valid: true, String: userData.ID},
				SpotifyName:     sql.NullString{Valid: true, String: userData.User.DisplayName},
				SpotifyImageUrl: sql.NullString{Valid: true, String: userData.Images[0].URL},
			})
			if err != nil {
				fmt.Printf("Could not update spotify data for user %s: %s\n", user.ID, err)
			}
		}
	}
}
