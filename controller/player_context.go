package controller

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/zmb3/spotify/v2"
)

func (c *Controller) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRequestContext(ctx, r)
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

	playlistID := spotify.ID(id)
	playlist, err := client.GetPlaylist(ctx, playlistID)
	if err != nil {
		log.Printf("error getting playlist: %s", err)
		requests.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(playlist)
}

func (c *Controller) GetAlbum(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRequestContext(ctx, r)
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

	reqCtx, err := getRequestContext(ctx, r)
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
