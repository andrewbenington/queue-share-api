package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/constants"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/room"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

func (c *Controller) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req room.InsertRoomParams
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorBadRequest))
		return
	}
	ctx := r.Context()

	authenticator := spotifyauth.New(spotifyauth.WithScopes(auth.SpotifyScopes...))
	httpClient := authenticator.Client(ctx, &oauth2.Token{
		AccessToken: req.AccessToken,
	})
	spClient := spotify.New(httpClient)
	user, err := spClient.CurrentUser(ctx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}

	req.HostID = user.ID
	req.HostName = user.DisplayName

	newRoom, err := db.Service().RoomStore.Insert(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error inserting room: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}

	body, err := json.MarshalIndent(newRoom, "", " ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling room: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func (c *Controller) GetRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code, password := room.ParametersFromRequest(r)
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorBadRequest))
		return
	}
	status, _, err := client.DecryptRoomToken(r.Context(), code, password)
	if status == http.StatusNotFound {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	if err != nil {
		w.WriteHeader(status)
		if status == http.StatusUnauthorized {
			_, _ = w.Write(MarshalErrorBody(constants.ErrorPassword))
		} else {
			fmt.Fprintf(os.Stderr, "error validating password: %s\n", err)
		}
		return
	}
	room, err := db.Service().RoomStore.GetByCode(ctx, code)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(MarshalErrorBody("Room not found"))
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching room: %s", err)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}
	body, err := json.MarshalIndent(room, "", " ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling room: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}
	w.Write(body)
}
