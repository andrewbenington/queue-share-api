package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/andrewbenington/go-spotify/constants"
	"github.com/andrewbenington/go-spotify/db"
	"github.com/gorilla/mux"
)

type CreateRoomRequest struct {
	Name string `json:"name"`
}

func (c *Controller) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req CreateRoomRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(constants.ErrorBadRequest))
		return
	}

	newRoom, err := db.Service().RoomStore.Insert(context.Background(), req.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error inserting room: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(constants.ErrorInternal))
		return
	}

	body, err := json.MarshalIndent(newRoom, "", " ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling room: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(constants.ErrorInternal))
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
		w.Write([]byte(constants.ErrorInternal))
		return
	}
	body, err := json.MarshalIndent(rooms, "", " ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling rooms: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(constants.ErrorInternal))
		return
	}
	w.Write(body)
}

func (c *Controller) GetRoom(w http.ResponseWriter, r *http.Request) {
	urlVars := mux.Vars(r)
	code, ok := urlVars["code"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(constants.ErrorBadRequest))
		return
	}
	room, err := db.Service().RoomStore.GetByCode(context.Background(), code)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching room: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(constants.ErrorInternal))
		return
	}
	body, err := json.MarshalIndent(room, "", " ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error marshalling room: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(constants.ErrorInternal))
		return
	}
	w.Write(body)
}
