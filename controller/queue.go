package controller

import (
	"encoding/json"
	"net/http"

	"github.com/andrewbenington/go-spotify/queue"
	"github.com/gorilla/mux"
)

var (
	SongIDMissingError, _ = json.MarshalIndent(ErrorResponse{
		Error: "No song specified",
	}, "", " ")
)

func (c *Controller) GetQueue(w http.ResponseWriter, r *http.Request) {
	currrentQueue, err := queue.GetUserQueue(c.Client)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err))
		return
	}
	responseBytes, err := json.MarshalIndent(currrentQueue, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err))
		return
	}
	_, _ = w.Write(responseBytes)
}

func (c *Controller) PushToQueue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	songID := vars["song"]
	if songID == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("no song specified"))
		return
	}
	err := queue.PushToUserQueue(c.Client, songID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err))
		return
	}
	currrentQueue, err := queue.GetUserQueue(c.Client)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err))
		return
	}
	responseBytes, err := json.MarshalIndent(currrentQueue, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err))
		return
	}
	_, _ = w.Write(responseBytes)
}
