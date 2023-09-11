package controller

import (
	"encoding/json"
	"net/http"

	"github.com/andrewbenington/go-spotify/client"
	"github.com/andrewbenington/go-spotify/queue"
	"github.com/gorilla/mux"
)

var (
	SongIDMissingError, _ = json.MarshalIndent(ErrorResponse{
		Error: "No song specified",
	}, "", " ")
)

func (c *Controller) GetQueue(w http.ResponseWriter, r *http.Request) {
	status, spClient, err := client.FromRequest(r)
	if err != nil {
		w.WriteHeader(status)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	currrentQueue, err := queue.GetUserQueue(spClient)
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
	status, spClient, err := client.FromRequest(r)
	if err != nil {
		w.WriteHeader(status)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	err = queue.PushToUserQueue(spClient, songID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	currrentQueue, err := queue.GetUserQueue(spClient)
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
