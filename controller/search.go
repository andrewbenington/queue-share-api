package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/zmb3/spotify/v2"
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

	json.NewEncoder(w).Encode(results)
}

func (c *Controller) GetArtistsByURIs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, authenticatedAsUser := ctx.Value(auth.UserContextKey).(string)
	if !authenticatedAsUser {
		requests.RespondAuthError(w)
		return
	}

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

	urisParam := r.URL.Query().Get("uris")
	if urisParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(SearchMissingError)
	}

	uris := strings.Split(urisParam, ",")
	ids := lo.Map(uris, service.IDFromURIMustIdx)
	artists, err := service.GetArtists(ctx, spClient, ids)
	if err != nil {
		requests.RespondWithError(w, code, err.Error())
		return
	}

	artistData := lo.MapEntries(artists, func(k string, v spotify.FullArtist) (string, db.ArtistData) {
		return k, service.ArtistDataFromFullArtist(v)
	})
	json.NewEncoder(w).Encode(artistData)
}

func (c *Controller) GetTracksByURIs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, authenticatedAsUser := ctx.Value(auth.UserContextKey).(string)
	if !authenticatedAsUser {
		requests.RespondAuthError(w)
		return
	}

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

	urisParam := r.URL.Query().Get("uris")
	if urisParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write(SearchMissingError)
	}

	uris := strings.Split(urisParam, ",")
	ids := lo.Map(uris, service.IDFromURIMustIdx)
	ids = lo.Uniq(ids)
	tracks, err := service.GetTracks(ctx, spClient, ids)
	if err != nil {
		requests.RespondWithError(w, code, err.Error())
		return
	}

	json.NewEncoder(w).Encode(tracks)
}
