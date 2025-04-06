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

func (c *Controller) GetPlaylist(w http.ResponseWriter, r *http.Request) {

	log.Printf("GetPlaylist")
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		fmt.Println(err)
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	id := (r.URL.Query().Get("playlist_id"))
	if id == "" {
		requests.RespondWithError(w, http.StatusBadRequest, "ID is empty")
		return
	}

	log.Printf("getting playlist '%s'", id)

	playlistID := spotify.ID(id)
	playlist, err := spClient.GetPlaylist(ctx, playlistID)
	if err != nil {
		log.Printf("error getting playlist: %s", err)
		requests.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(playlist)
}

func (c *Controller) GetAlbum(w http.ResponseWriter, r *http.Request) {
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

	id := (r.URL.Query().Get("id"))
	if id == "" {
		requests.RespondBadRequest(w)
		return
	}

	albumID := spotify.ID(id)
	album, err := client.GetAlbum(ctx, albumID)
	if err != nil {
		log.Printf("error getting album: %s", err)
		requests.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(album)
}

func (c *Controller) GetArtist(w http.ResponseWriter, r *http.Request) {
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

	id := (r.URL.Query().Get("id"))
	if id == "" {
		requests.RespondBadRequest(w)
		return
	}

	artistID := spotify.ID(id)
	artist, err := client.GetArtist(ctx, artistID)
	if err != nil {
		log.Printf("error getting artist: %s", err)
		requests.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(artist)
}
