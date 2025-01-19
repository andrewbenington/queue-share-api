package controller

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
)

type MixBuilder struct {
	ArtistIDs   []string `json:"artist_ids"`
	AlbumIDs    []string `json:"album_ids"`
	PlaylistIDs []string `json:"playlist_ids"`
}

func (c *StatsController) AddMixToQueue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		fmt.Println(err)
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	var req MixBuilder
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Println(err)
		requests.RespondBadRequest(w)
		return
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	playlistTrackURIs := []string{}
	for _, playlistID := range req.PlaylistIDs {
		playlist, err := spClient.GetPlaylist(ctx, spotify.ID(playlistID))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		perm := rand.Perm(len(playlist.Tracks.Tracks))
		for _, v := range perm[:10] {
			track := playlist.Tracks.Tracks[v]
			if strings.HasPrefix(string(track.Track.URI), "spotify:local") {
				continue
			}
			playlistTrackURIs = append(playlistTrackURIs, string(track.Track.URI))
		}
	}

	transaction, err := db.Service().DB.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer transaction.Commit()

	albumURIs := lo.Map(
		req.AlbumIDs,
		func(id string, _ int) string {
			if id == "" {
				fmt.Printf("MISSING ALBUM")
			}
			return fmt.Sprintf("spotify:album:%s", id)
		})

	albumStreams, err := db.New(transaction).HistoryGetAlbumStreams(ctx, db.HistoryGetAlbumStreamsParams{
		UserID:    userUUID,
		AlbumUris: albumURIs,
	})
	if err != nil {
		fmt.Println(err, "could not get album streams")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	artistURIs := lo.Map(
		req.ArtistIDs,
		func(id string, _ int) string {
			if id == "" {
				fmt.Printf("MISSING ARTIST")
			}
			return fmt.Sprintf("spotify:artist:%s", id)
		})

	artistStreams, err := db.New(transaction).HistoryGetRecentArtistStreams(ctx, db.HistoryGetRecentArtistStreamsParams{
		UserID:     userUUID,
		ArtistUris: artistURIs,
	})
	if err != nil {
		fmt.Println(err, "could not get artist streams")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	allTrackURIsSet := map[string]bool{}
	for _, stream := range albumStreams {
		allTrackURIsSet[stream.SpotifyTrackUri] = true
	}

	for _, stream := range artistStreams {
		allTrackURIsSet[stream.SpotifyTrackUri] = true
	}
	fmt.Printf("ARTIST: %+v\n", artistStreams)

	for _, uri := range playlistTrackURIs {
		allTrackURIsSet[uri] = true
	}

	allTrackURIs := lo.Keys(allTrackURIsSet)
	fmt.Println("ALL", allTrackURIs)
	shuffledTrackURIs := make([]string, len(allTrackURIsSet))
	perm := rand.Perm(len(allTrackURIsSet))
	for i, v := range perm {
		shuffledTrackURIs[v] = allTrackURIs[i]
	}

	for _, uri := range shuffledTrackURIs {
		err = service.PushToUserQueue(r.Context(), spClient, service.IDFromURIMust(uri))
		if err != nil && strings.Contains(err.Error(), "No active device found") {
			requests.RespondWithError(w, http.StatusBadRequest, "Host is not playing music")
			return
		}
		if err != nil {
			fmt.Println(err, "could not enqueue")
			requests.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}
