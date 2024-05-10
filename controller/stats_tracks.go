package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrewbenington/queue-share-api/auth"
	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/spotify"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type TopTracksResponse struct {
	Rankings  []*history.MonthTopTracks `json:"rankings"`
	TrackData map[string]db.TrackData   `json:"track_data"`
}

func (c *Controller) GetTopTracksByMonth(w http.ResponseWriter, r *http.Request) {
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

	filter := getFilterParams(r)

	transaction := db.Service().DB

	rankingResults, code, err := history.TrackStreamRankingsByMonth(ctx, transaction, userID, filter)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	trackIDs := map[string]bool{}
	for _, result := range rankingResults {
		for _, track := range result.Tracks {
			trackIDs[track.ID] = true
		}
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	trackResults, err := spotify.GetTracks(ctx, spClient, maps.Keys(trackIDs))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := TopTracksResponse{
		Rankings:  rankingResults,
		TrackData: trackResults,
	}

	json.NewEncoder(w).Encode(response)
}

func GetTopItemsByMonth(w http.ResponseWriter, r *http.Request) {
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

	filter := getFilterParams(r)

	transaction := db.Service().DB

	rankingResults, code, err := history.TrackStreamRankingsByMonth(ctx, transaction, userID, filter)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	trackIDs := map[string]bool{}

	for _, result := range rankingResults {
		for _, track := range result.Tracks {
			trackIDs[track.ID] = true
		}
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	trackResults, err := spotify.GetTracks(ctx, spClient, maps.Keys(trackIDs))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := TopTracksResponse{
		Rankings:  rankingResults,
		TrackData: trackResults,
	}

	json.NewEncoder(w).Encode(response)
}
