package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/constants"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/gorilla/mux"
)

var (
	SongIDMissingError, _ = json.MarshalIndent(requests.ErrorResponse{
		Error: "No song specified",
	}, "", " ")
)

func (c *Controller) GetQueue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Guest {
		requests.RespondWithRoomAuthError(w, int(reqCtx.PermissionLevel))
		return
	}

	status, spClient, err := client.ForRoom(ctx, reqCtx.Room.Code)
	if err != nil {
		requests.RespondWithError(w, status, err.Error())
		return
	}
	currentQueue, err := spotify.GetUserQueue(r.Context(), spClient)
	if err != nil {
		log.Printf("Error getting user queue: %s", err)
		requests.RespondInternalError(w)
		return
	}

	err = spotify.UpdateUserPlayback(r.Context(), spClient, currentQueue)
	if err != nil {
		log.Printf("Error updating user playback: %s", err)
		requests.RespondInternalError(w)
		return
	}

	status, errMessage := addGuestsAndMembersToTracks(r.Context(), reqCtx.Room.ID, currentQueue)
	if status != http.StatusOK {
		requests.RespondWithError(w, status, errMessage)
		return
	}

	responseBytes, err := json.MarshalIndent(currentQueue, "", " ")
	if err != nil {
		requests.RespondInternalError(w)
		return
	}
	_, _ = w.Write(responseBytes)
}

func (c *Controller) PushToQueue(w http.ResponseWriter, r *http.Request) {
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

	vars := mux.Vars(r)
	songID, ok := vars["song"]
	if !ok || songID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no song specified"))
		return
	}
	status, spClient, err := client.ForRoom(ctx, reqCtx.Room.Code)
	if err != nil {
		requests.RespondWithError(w, status, err.Error())
		return
	}
	err = spotify.PushToUserQueue(r.Context(), spClient, songID)
	if err != nil && strings.Contains(err.Error(), "No active device found") {
		requests.RespondWithError(w, http.StatusBadRequest, "Host is not playing music")
		return
	}
	if err != nil {
		log.Printf("Error pushing to user queue: %s", err)
		requests.RespondInternalError(w)
		return
	}

	if reqCtx.UserID != "" {
		err = db.Service().RoomStore.SetQueueTrackUser(ctx, reqCtx.Room.Code, songID, reqCtx.UserID)
		if err != nil {
			requests.RespondWithDBError(w, err)
			return
		}
	} else if reqCtx.GuestID != "" {
		err = db.Service().RoomStore.SetQueueTrackGuest(ctx, reqCtx.Room.Code, songID, reqCtx.GuestID)
		if err != nil {
			requests.RespondWithDBError(w, err)
			return
		}
	}

	time.Sleep(500 * time.Millisecond)

	currentQueue, err := spotify.GetUserQueue(r.Context(), spClient)
	if err != nil {
		requests.RespondInternalError(w)
		return
	}

	err = spotify.UpdateUserPlayback(r.Context(), spClient, currentQueue)
	if err != nil {
		requests.RespondInternalError(w)
		return
	}

	status, errMessage := addGuestsAndMembersToTracks(r.Context(), reqCtx.Room.ID, currentQueue)
	if status != http.StatusOK {
		requests.RespondWithError(w, status, errMessage)
		return
	}

	responseBytes, err := json.MarshalIndent(currentQueue, "", " ")
	if err != nil {
		requests.RespondInternalError(w)
		return
	}
	_, _ = w.Write(responseBytes)
}

func addGuestsAndMembersToTracks(ctx context.Context, roomID string, q *spotify.CurrentQueue) (statusCode int, errMessage string) {
	guestTracks, err := db.Service().RoomStore.GetQueueTrackAddedBy(ctx, roomID)
	if err == sql.ErrNoRows {
		return http.StatusNotFound, constants.ErrorNotFound
	}
	if err != nil {
		return http.StatusInternalServerError, constants.ErrorInternal
	}

	tracks := make(map[string]*spotify.TrackInfo)

	tracks[q.CurrentlyPlaying.ID] = &q.CurrentlyPlaying
	for i := range q.Queue {
		if _, ok := tracks[q.Queue[i].ID]; !ok {
			tracks[q.Queue[i].ID] = &q.Queue[i]
		}
	}

	for _, gt := range guestTracks {
		queueTrack, ok := tracks[gt.TrackID]
		if ok && !gt.Played && gt.Timestamp.After(queueTrack.AddedAt) {
			queueTrack.AddedBy = gt.AddedBy
			queueTrack.AddedAt = gt.Timestamp
			if queueTrack.ID == q.CurrentlyPlaying.ID {
				db.Service().RoomStore.MarkTracksAsPlayedSince(ctx, roomID, gt.Timestamp)
			}
		}
	}

	return http.StatusOK, ""
}
