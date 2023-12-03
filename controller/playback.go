package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/zmb3/spotify/v2"
)

func (c *Controller) Pause(w http.ResponseWriter, r *http.Request) {
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

	err = client.Pause(ctx)
	if err != nil {
		log.Printf("error pausing song: %s", err)
		requests.RespondWithError(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) Play(w http.ResponseWriter, r *http.Request) {
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

	var playOptions *spotify.PlayOptions
	deviceID := (spotify.ID)(r.URL.Query().Get("device_id"))
	playlistID := (spotify.ID)(r.URL.Query().Get("playlist_id"))
	if deviceID != "" && playlistID != "" {
		pc := (spotify.URI)(fmt.Sprintf("spotify:playlist:%s", playlistID))
		playOptions = &spotify.PlayOptions{
			DeviceID:        &deviceID,
			PlaybackContext: &pc,
		}
	}

	err = client.PlayOpt(ctx, playOptions)
	if err != nil {
		log.Printf("error playing song: %s", err)
		requests.RespondWithError(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) Next(w http.ResponseWriter, r *http.Request) {
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

	err = client.Next(ctx)
	if err != nil {
		log.Printf("error playing next song: %s", err)
		requests.RespondWithError(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) Previous(w http.ResponseWriter, r *http.Request) {
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

	err = client.Previous(ctx)
	if err != nil {
		log.Printf("error playing previous song: %s", err)
		requests.RespondWithError(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) Devices(w http.ResponseWriter, r *http.Request) {
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

	devices, err := client.PlayerDevices(ctx)
	if err != nil {
		log.Printf("error getting devices: %s", err)
		requests.RespondWithError(w, status, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(devices)
}

func (c *Controller) SetVolume(w http.ResponseWriter, r *http.Request) {
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

	percentParam := (r.URL.Query().Get("percent"))
	percent, err := strconv.Atoi(percentParam)
	if err != nil {
		requests.RespondBadRequest(w)
		return
	}

	err = client.Volume(ctx, percent)
	if err != nil {
		log.Printf("error setting volume: %s", err)
		requests.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) GetPlayback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Guest {
		requests.RespondAuthError(w)
		return
	}

	status, client, err := client.ForRoom(ctx, reqCtx.Room.Code)
	if err != nil {
		requests.RespondWithError(w, status, err.Error())
		return
	}

	state, err := client.PlayerState(ctx)
	if err != nil {
		log.Printf("error getting player state: %s", err)
		requests.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	json.NewEncoder(w).Encode(state)
}
