package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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

	time.Sleep(500 * time.Millisecond)
	c.GetQueue(w, r)
}

type PlayBody struct {
	ContextURI string `json:"context_uri"`
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

	time.Sleep(500 * time.Millisecond)
	c.GetQueue(w, r)
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

	time.Sleep(500 * time.Millisecond)
	c.GetQueue(w, r)
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

	time.Sleep(500 * time.Millisecond)
	c.GetQueue(w, r)
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
