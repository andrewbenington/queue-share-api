package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/gorilla/mux"
)

var (
	SongIDMissingError, _ = json.MarshalIndent(ErrorResponse{
		Error: "No song specified",
	}, "", " ")
)

func (c *Controller) GetQueue(w http.ResponseWriter, r *http.Request) {
	status, spClient, err := client.ForRoom(r)
	if err != nil {
		w.WriteHeader(status)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	currrentQueue, err := spotify.GetUserQueue(r.Context(), spClient)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	responseBytes, err := json.MarshalIndent(currrentQueue, "", " ")
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
	time.Sleep(500 * time.Millisecond)
	currrentQueue, err := spotify.GetUserQueue(r.Context(), spClient)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	responseBytes, err := json.MarshalIndent(currrentQueue, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	_, _ = w.Write(responseBytes)
}
