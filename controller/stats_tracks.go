package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
	"golang.org/x/exp/maps"
)

type TopTracksResponse struct {
	Rankings  []*history.MonthTopTracks `json:"rankings"`
	TrackData map[string]db.TrackData   `json:"track_data"`
}

func (c *StatsController) GetTopTracksByMonth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	fmt.Println(userUUID)

	filter := getFilterParams(r)
	filter.Max = 30

	transaction := db.Service().DB

	rankingResults, code, err := history.TrackStreamRankingsByMonth(ctx, transaction, userUUID, filter)
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

	trackResults, err := service.GetTracks(ctx, spClient, maps.Keys(trackIDs))
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

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}

	filter := getFilterParams(r)

	transaction := db.Service().DB

	rankingResults, code, err := history.TrackStreamRankingsByMonth(ctx, transaction, userUUID, filter)
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

	trackResults, err := service.GetTracks(ctx, spClient, maps.Keys(trackIDs))
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

type Stream struct {
	Timestamp       time.Time `json:"timestamp"`
	MsPlayed        int       `json:"ms_played"`
	TrackName       string    `json:"track_name"`
	AlbumName       string    `json:"album_name"`
	SpotifyTrackUri string    `json:"spotify_track_uri"`
	SpotifyAlbumUri string    `json:"spotify_album_uri"`
	ISRC            *string   `json:"isrc,omitempty"`
}
type TrackStatsResponse struct {
	Track   *db.TrackData `json:"track"`
	Streams []*Stream     `json:"streams"`
}

func (c *StatsController) GetTrackStatsByURI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	trackURI := r.URL.Query().Get("spotify_uri")
	trackID, err := service.IDFromURI(trackURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	track, err := service.GetTrack(ctx, spClient, trackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows, err := history.AllTrackStreamsByURI(
		ctx,
		db.Service().DB,
		userUUID,
		trackURI,
		getFilterParams(r),
	)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	streams := []*Stream{}
	for _, row := range rows {
		streams = append(streams, &Stream{
			Timestamp:       row.Timestamp,
			TrackName:       row.TrackName,
			AlbumName:       row.AlbumName,
			MsPlayed:        int(row.MsPlayed),
			SpotifyTrackUri: row.SpotifyTrackUri,
			SpotifyAlbumUri: row.SpotifyAlbumUri.String,
			ISRC:            row.Isrc,
		})
	}

	response := TrackStatsResponse{
		Streams: streams,
		Track:   track,
	}
	json.NewEncoder(w).Encode(response)
}

func (c *StatsController) GetTrackRankingsByURI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, err.Error())
		return
	}
	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	trackURI := r.URL.Query().Get("spotify_uri")
	trackID, err := service.IDFromURI(trackURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	track, err := service.GetTrack(ctx, spClient, trackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filter := getFilterParams(r)
	filter.Max = 30
	allRankings, responseCode, err := history.TrackStreamRankingsByMonth(ctx, db.Service().DB, userUUID, filter)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	rankings := []MonthRanking{}

	for _, monthRankings := range allRankings {
		for i, trackPlays := range monthRankings.Tracks {
			if trackPlays.ID == string(track.ID) {
				ranking := MonthRanking{
					Year:     monthRankings.Year,
					Month:    monthRankings.Month,
					Position: i + 1,
				}
				rankings = append(rankings, ranking)
			}
		}
	}

	json.NewEncoder(w).Encode(rankings)
}
