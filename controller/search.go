package controller

import (
	"encoding/json"
	"net/http"

	"github.com/andrewbenington/go-spotify/client"
	"github.com/andrewbenington/go-spotify/search"
)

var (
	SearchMissingError, _ = json.MarshalIndent(ErrorResponse{
		Error: "No search term present",
	}, "", " ")
)

func (c *Controller) Search(w http.ResponseWriter, r *http.Request) {
	status, client, err := client.FromRequest(r)
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
	results, err := search.SearchSongs(client, term)
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
