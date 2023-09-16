package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/andrewbenington/go-spotify/client"
	"github.com/andrewbenington/go-spotify/constants"
	"github.com/andrewbenington/go-spotify/db"
	"github.com/andrewbenington/go-spotify/room"
)

type CreateRoomRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (c *Controller) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req room.InsertRoomParams
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorBadRequest))
		return
	}

	newRoom, err := db.Service().RoomStore.Insert(r.Context(), req)
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

func (c *Controller) GetAllRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := db.Service().RoomStore.All(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching rooms: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}
	body, err := json.MarshalIndent(rooms, "", " ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling rooms: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(constants.ErrorInternal))
		return
	}
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
	if err != nil {
		w.WriteHeader(status)
		if status == http.StatusUnauthorized {
			_, _ = w.Write(MarshalErrorBody(constants.ErrorPassword))
		} else {
			fmt.Fprintf(os.Stderr, "error validating password: %s", err)
		}
		return
	}
	room, err := db.Service().RoomStore.GetByCode(ctx, code)

	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
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
