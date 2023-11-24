package controller

import (
	"encoding/json"
	"net/http"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/spotify"
)

var (
	SearchMissingError, _ = json.MarshalIndent(requests.ErrorResponse{
		Error: "No search term present",
	}, "", " ")
)

func (c *Controller) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRequestContext(ctx, r)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	if reqCtx.PermissionLevel < Guest {
		requests.RespondAuthError(w)
		return
	}

	status, client, err := client.ForRoom(ctx, reqCtx.Room.Code)
	if err != nil {
		requests.RespondWithError(w, status, err.Error())
		return
	}

	term := r.URL.Query().Get("q")
	if term == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(SearchMissingError)
	}

	results, err := spotify.SearchSongs(r.Context(), client, term)
	if err != nil {
		requests.RespondInternalError(w)
		return
	}

	responseBytes, err := json.MarshalIndent(results, "", " ")
	if err != nil {
		requests.RespondInternalError(w)
		return
	}
	_, _ = w.Write(responseBytes)
}
