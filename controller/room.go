package controller

import (
	"context"
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

	body, err := json.Marshal(room.RoomResponse{Room: newRoom})
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
	code, guestID, password := room.ParametersFromRequest(r)
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorBadRequest))
		return
	}
	status, errMsg := validateRoomPass(ctx, code, password)
	if status != http.StatusOK {
		w.WriteHeader(status)
		_, _ = w.Write(MarshalErrorBody(errMsg))
		return
	}

	rm, err := db.Service().RoomStore.GetByCode(ctx, code)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(MarshalErrorBody("Room not found"))
		return
	}
	if err != nil {
		log.Printf("error fetching room: %s\n", err)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}

	resp := room.RoomResponse{
		Room: rm,
	}

	if guestID != "" {
		guestName, err := db.Service().RoomStore.GetGuestName(ctx, rm.ID, guestID)
		if err != nil {
			log.Printf("error finding guest name: %s\n", err)
		} else {
			resp.Guest = &room.Guest{
				ID:   guestID,
				Name: guestName,
			}
		}
	}

	body, err := json.Marshal(resp)
	if err != nil {
		log.Printf("error marshalling room response: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}
	w.Write(body)
}

func (*Controller) AddGuest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code, guestID, password := room.ParametersFromRequest(r)
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorBadRequest))
		return
	}

	status, errMsg := validateRoomPass(ctx, code, password)
	if status != http.StatusOK {
		w.WriteHeader(status)
		_, _ = w.Write(MarshalErrorBody(errMsg))
		return
	}

	var req room.InsertGuestRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorBadRequest))
		return
	}

	var guest *room.Guest
	if guestID != "" {
		guest, err = db.Service().RoomStore.InsertGuestWithID(ctx, code, req.Name, guestID)
	} else {
		guest, err = db.Service().RoomStore.InsertGuest(ctx, code, req.Name)
	}

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorNotFound))
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}

	body, err := json.Marshal(guest)
	if err != nil {
		log.Printf("error marshalling room guest: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(body)
}

func validateRoomPass(ctx context.Context, code string, password string) (status int, errorMessage string) {
	valid, err := db.Service().RoomStore.ValidatePassword(ctx, code, password)
	if err == sql.ErrNoRows {
		return http.StatusNotFound, constants.ErrorNotFound
	}
	if err != nil {
		log.Printf("error authenticating room: %s\n", err)
		return http.StatusInternalServerError, constants.ErrorInternal
	}
	if !valid {
		return http.StatusForbidden, constants.ErrorPassword
	}
	return http.StatusOK, ""
}
