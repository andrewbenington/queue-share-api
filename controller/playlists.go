package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/zmb3/spotify/v2"
)

func (c *Controller) RoomPlaylists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRoomRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Moderator {
		requests.RespondAuthError(w)
		return
	}

	status, client, err := client.ForRoom(ctx, reqCtx.Room.Code)
	if err != nil {
		requests.RespondWithError(w, status, err.Error())
		return
	}

	playlists, err := client.CurrentUsersPlaylists(ctx, spotify.Limit(40))
	if err != nil {
		log.Printf("error getting playlists: %s", err)
		requests.RespondWithError(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(playlists)
}

func (c *Controller) UserPlaylists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	status, client, err := client.ForUser(ctx, userUUID)
	if err != nil {
		requests.RespondWithError(w, status, err.Error())
		return
	}

	playlists, err := client.CurrentUsersPlaylists(ctx, spotify.Limit(40))
	if err != nil {
		log.Printf("error getting playlists: %s", err)
		requests.RespondWithError(w, status, err.Error())
		return
	}
	allPlaylists := playlists

	offset := 40
	for playlists.Next != "" {
		fmt.Println("Next:", offset)
		playlists, err = client.CurrentUsersPlaylists(ctx, spotify.Limit(40), spotify.Offset(offset))
		if err != nil {
			log.Printf("error getting playlists: %s", err)
			requests.RespondWithError(w, status, err.Error())
			return
		}
		allPlaylists.Playlists = append(allPlaylists.Playlists, playlists.Playlists...)
		offset += len(playlists.Playlists)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(allPlaylists)
}
