package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/spotify"
)

var (
	SearchMissingError, _ = json.MarshalIndent(ErrorResponse{
		Error: "No search term present",
	}, "", " ")
)

func (c *Controller) Search(w http.ResponseWriter, r *http.Request) {
	status, client, err := client.ForRoom(r)
	if err != nil {
		w.WriteHeader(status)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	term := r.URL.Query().Get("q")
	if term == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(SearchMissingError)
	}
	fmt.Println(term)
	results, err := spotify.SearchSongs(r.Context(), client, term)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	responseBytes, err := json.MarshalIndent(results, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(MarshalErrorBody(err.Error()))
		return
	}
	_, _ = w.Write(responseBytes)
}
