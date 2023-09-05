package playlist

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/andrewbenington/go-spotify/auth"
	"github.com/zmb3/spotify/v2"
)

func ListPlaylists() error {
	client := auth.Client()
	if client == nil {
		return errors.New("not authenticated")
	}
	playlists, err := client.CurrentUsersPlaylists(context.Background(), spotify.Limit(50))
	if err != nil {
		return fmt.Errorf("could not get playlists: %w", err)
	}
	listPlaylistsFromPage(playlists)
	for playlists.Next != "" {
		fmt.Println(playlists.Next)
		playlists, err = client.CurrentUsersPlaylists(context.Background(), spotify.Limit(50), spotify.After(playlists.Next))
		if err != nil {
			return fmt.Errorf("could not get playlists page %s: %w", playlists.Next, err)
		}
	}

	return nil
}

func listPlaylistsFromPage(page *spotify.SimplePlaylistPage) {
	client := auth.Client()
	user, _ := client.CurrentUser(context.Background())
	for _, p := range page.Playlists {
		if p.Owner.ID != user.ID {
			continue
		}
		fmt.Println(p.Name)
		if p.Description != "" {
			fmt.Printf("\t%s\n", p.Description)
			continue
		}
		details, err := client.GetPlaylist(context.Background(), p.ID)
		if err != nil {
			fmt.Printf("could not get details for %s: %s\n", p.Name, err)
			continue
		}
		earliest := time.Now()
		for _, t := range details.Tracks.Tracks {
			addedTime, err := time.Parse(spotify.TimestampLayout, t.AddedAt)
			if err == nil && addedTime.Before(earliest) {
				earliest = addedTime
			}
		}
		fmt.Printf("\t%s\n", earliest.Local().Format("January 2, 2006"))
		err = client.ChangePlaylistDescription(context.Background(), p.ID, earliest.Local().Format("January 2, 2006"))
		if err != nil {
			fmt.Printf("could not update %s: %s\n", p.Name, err)
			continue
		}

	}
}
