package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/constants"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/room"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/gorilla/mux"
)

var (
	SongIDMissingError, _ = json.MarshalIndent(requests.ErrorResponse{
		Error: "No song specified",
	}, "", " ")
)

func (c *Controller) GetQueue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRoomRequestContext(ctx, r)
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
	currentQueue, err := service.GetUserQueue(r.Context(), spClient)
	if err != nil {
		log.Printf("Error getting user queue: %s", err)
		requests.RespondInternalError(w)
		return
	}

	err = service.UpdateUserPlayback(r.Context(), spClient, currentQueue)
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

func (c *Controller) PushToRoomQueue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRoomRequestContext(ctx, r)
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
	err = service.PushToUserQueue(r.Context(), spClient, songID)
	if err != nil && strings.Contains(err.Error(), "No active device found") {
		requests.RespondWithError(w, http.StatusBadRequest, "Host is not playing music")
		return
	}
	if err != nil {
		log.Printf("Error pushing to user queue: %s", err)
		requests.RespondInternalError(w)
		return
	}

	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}

	if reqCtx.UserID != "" {
		err = room.SetQueueTrackUser(ctx, tx, reqCtx.Room.Code, songID, reqCtx.UserID)
		if err != nil {
			requests.RespondWithDBError(w, err)
			return
		}
	} else if reqCtx.GuestID != "" {
		err = room.SetQueueTrackGuest(ctx, tx, reqCtx.Room.Code, songID, reqCtx.GuestID)
		if err != nil {
			requests.RespondWithDBError(w, err)
			return
		}
	}

	time.Sleep(500 * time.Millisecond)

	currentQueue, err := service.GetUserQueue(r.Context(), spClient)
	if err != nil {
		requests.RespondInternalError(w)
		return
	}

	err = service.UpdateUserPlayback(r.Context(), spClient, currentQueue)
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

	err = tx.Commit(ctx)
	if err != nil {
		http.Error(w, "Error committing DB transaction", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(responseBytes)
}

func (c *Controller) PushToUserQueue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	trackURI := r.URL.Query().Get("track")
	trackID, err := service.IDFromURI(trackURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad track uri: %s", err), http.StatusBadRequest)
		return
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	err = service.PushToUserQueue(r.Context(), spClient, trackID)
	if err != nil && strings.Contains(err.Error(), "No active device found") {
		requests.RespondWithError(w, http.StatusBadRequest, "Host is not playing music")
		return
	}
	if err != nil {
		log.Printf("Error pushing to user queue: %s", err)
		requests.RespondInternalError(w)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func addGuestsAndMembersToTracks(ctx context.Context, roomID string, q *service.CurrentQueue) (statusCode int, errMessage string) {
	tx, err := db.Service().BeginTx(ctx)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}

	guestTracks, err := room.GetQueueTrackAddedBy(ctx, tx, roomID)
	if err == sql.ErrNoRows {
		return http.StatusNotFound, constants.ErrorNotFound
	}
	if err != nil {
		return http.StatusInternalServerError, constants.ErrorInternal
	}

	tracks := make(map[string]*service.TrackInfo)

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
				room.MarkTracksAsPlayedSince(ctx, tx, roomID, gt.Timestamp)
			}
		}
	}

	return http.StatusOK, ""
}
