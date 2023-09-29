package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/constants"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/room"
)

func (c *Controller) CreateRoom(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(auth.UserContextKey).(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorNotAuthenticated))
		return
	}

	var req room.InsertRoomParams
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorBadRequest))
		return
	}

	req.HostID = userID

	newRoom, err := db.Service().RoomStore.Insert(r.Context(), req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error inserting room: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}

	user, err := db.Service().UserStore.GetByID(r.Context(), userID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "get room host after create: %s", err)
	} else {
		newRoom.Host = *user
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
	valid, err := db.Service().RoomStore.ValidatePassword(ctx, code, password)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorNotFound))
		return
	}
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}
	if !valid {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorPassword))
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
