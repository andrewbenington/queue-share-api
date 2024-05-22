package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/google/uuid"
)

var (
	SearchMissingError, _ = json.MarshalIndent(requests.ErrorResponse{
		Error: "No search term present",
	}, "", " ")
)

func (c *Controller) SearchTracksFromRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	reqCtx, err := getRoomRequestContext(ctx, r)
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
	results, err := service.SearchTracks(r.Context(), client, term)
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

func (c *Controller) SearchTracksByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, authenticatedAsUser := ctx.Value(auth.UserContextKey).(string)
	if !authenticatedAsUser {
		requests.RespondAuthError(w)
		return
	}
	fmt.Println(userID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		requests.RespondWithError(w, code, err.Error())
		return
	}

	term := r.URL.Query().Get("q")
	if term == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(SearchMissingError)
	}
	results, err := service.SearchTracks(r.Context(), spClient, term)
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

func (c *Controller) SearchArtistsByUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, authenticatedAsUser := ctx.Value(auth.UserContextKey).(string)
	if !authenticatedAsUser {
		requests.RespondAuthError(w)
		return
	}
	fmt.Println(userID)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		requests.RespondWithError(w, code, err.Error())
		return
	}

	term := r.URL.Query().Get("q")
	if term == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(SearchMissingError)
	}
	results, err := service.SearchArtists(r.Context(), spClient, term)
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
