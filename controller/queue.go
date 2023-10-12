package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/constants"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/room"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/gorilla/mux"
)

var (
	SongIDMissingError, _ = json.MarshalIndent(ErrorResponse{
		Error: "No song specified",
	}, "", " ")
)

func (c *Controller) GetQueue(w http.ResponseWriter, r *http.Request) {
	roomCode, ok := mux.Vars(r)["code"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorNotFound))
		return
	}

	status, spClient, err := client.ForRoom(r)
	if err != nil {
		w.WriteHeader(status)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	currentQueue, err := spotify.GetUserQueue(r.Context(), spClient)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}

	err = spotify.UpdateUserPlayback(r.Context(), spClient, currentQueue)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}

	status, errMessage := addGuestsToTracks(r.Context(), roomCode, currentQueue)
	if status != http.StatusOK {
		w.WriteHeader(status)
		_, _ = w.Write(MarshalErrorBody(errMessage))
		return
	}

	responseBytes, err := json.MarshalIndent(currentQueue, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	_, _ = w.Write(responseBytes)
}

func (c *Controller) PushToQueue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songID, ok := vars["song"]
	if !ok || songID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no song specified"))
		return
	}
	status, spClient, err := client.ForRoom(r)
	if err != nil {
		w.WriteHeader(status)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	err = spotify.PushToUserQueue(r.Context(), spClient, songID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}

	code, guestID, _ := room.ParametersFromRequest(r)
	err = db.Service().RoomStore.SetQueueTrackGuest(r.Context(), code, songID, guestID)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorNotFound))
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}

	time.Sleep(500 * time.Millisecond)

	currentQueue, err := spotify.GetUserQueue(r.Context(), spClient)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}

	err = spotify.UpdateUserPlayback(r.Context(), spClient, currentQueue)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}

	status, errMessage := addGuestsToTracks(r.Context(), code, currentQueue)
	if status != http.StatusOK {
		w.WriteHeader(status)
		_, _ = w.Write(MarshalErrorBody(errMessage))
		return
	}

	responseBytes, err := json.MarshalIndent(currentQueue, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	_, _ = w.Write(responseBytes)
}

func addGuestsToTracks(ctx context.Context, roomCode string, q *spotify.CurrentQueue) (statusCode int, errMessage string) {
	guestTracks, err := db.Service().RoomStore.GetQueueTrackGuests(ctx, roomCode)
	if err == sql.ErrNoRows {
		return http.StatusNotFound, constants.ErrorNotFound
	}
	if err != nil {
		return http.StatusInternalServerError, constants.ErrorInternal
	}

	tracks := make(map[string]*spotify.TrackInfo)

	tracks[q.CurrentlyPlaying.ID] = &q.CurrentlyPlaying
	for i := range q.Queue {
		tracks[q.Queue[i].ID] = &q.Queue[i]
	}

	for _, gt := range guestTracks {
		queueTrack, ok := tracks[gt.TrackID]
		if ok {
			queueTrack.AddedBy = gt.GuestName
		}
	}

	return http.StatusOK, ""
}
