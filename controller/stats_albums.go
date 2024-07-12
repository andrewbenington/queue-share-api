package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrewbenington/queue-share-api/client"
	"github.com/andrewbenington/queue-share-api/db"
	"github.com/andrewbenington/queue-share-api/history"
	"github.com/andrewbenington/queue-share-api/requests"
	"github.com/andrewbenington/queue-share-api/service"
	"github.com/gorilla/mux"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/exp/maps"
)

type TopAlbumsResponse struct {
	Rankings  []*history.AlbumRankings     `json:"rankings"`
	AlbumData map[string]spotify.FullAlbum `json:"album_data"`
}

func (c *StatsController) GetTopAlbumsByTimeframe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	filter := getFilterParams(r)

	transaction, err := db.Service().DB.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer transaction.Commit()

	rankingResults, code, err := history.AlbumStreamRankingsByTimeframe(ctx, transaction, userUUID, filter, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	albumIDs := map[string]bool{}
	for _, result := range rankingResults {
		for _, album := range result.AlbumStreams {
			albumIDs[album.ID] = true
		}
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, err.Error(), code)
		return
	}

	albumResults, err := service.GetAlbums(ctx, spClient, maps.Keys(albumIDs))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := TopAlbumsResponse{
		Rankings:  rankingResults,
		AlbumData: albumResults,
	}
	json.NewEncoder(w).Encode(response)
}

type AlbumStatsResponse struct {
	Album      *spotify.FullAlbum      `json:"album"`
	Streams    []*Stream               `json:"streams"`
	Tracks     map[string]db.TrackData `json:"tracks"`
	TrackRanks []*history.TrackStreams `json:"track_ranks"`
}

func (c *StatsController) GetAlbumStatsByURI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}

	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get client: %s", err), code)
		return
	}

	albumURI := r.URL.Query().Get("spotify_uri")
	albumID, err := service.IDFromURI(albumURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad album uri: %s", err), http.StatusBadRequest)
		return
	}

	album, err := service.GetAlbum(ctx, spClient, albumID)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get album: %s", err), http.StatusInternalServerError)
		return
	}

	filter := getFilterParams(r)

	transaction, err := db.Service().DB.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer transaction.Commit()

	rows, err := history.AllAlbumStreamsByURI(
		ctx, transaction, userUUID, albumURI, filter,
	)
	if err != nil {
		requests.RespondWithDBError(w, err)
		return
	}

	filter.AlbumURI = &albumURI
	_, _, trackRanks, err := history.CalcTrackStreamsAndRanks(ctx, userUUID, filter, transaction, nil, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get track counts: %s", err), http.StatusInternalServerError)
		return
	}

	if trackRanks == nil {
		trackRanks = []*history.TrackStreams{}
	}

	trackIDs := map[string]bool{}
	for _, stream := range trackRanks {
		trackIDs[stream.ID] = true
	}

	streams := []*Stream{}
	for _, row := range rows {
		streams = append(streams, &Stream{
			Timestamp:       row.Timestamp,
			TrackName:       row.TrackName,
			AlbumName:       row.AlbumName,
			MsPlayed:        int(row.MsPlayed),
			SpotifyTrackUri: row.SpotifyTrackUri,
		})
		id, err := service.IDFromURI(row.SpotifyTrackUri)
		if err != nil {
			continue
		}
		trackIDs[id] = true
	}

	tracks, err := service.GetTracks(ctx, spClient, maps.Keys(trackIDs))
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get streamed tracks: %s", err), http.StatusInternalServerError)
		return
	}

	response := AlbumStatsResponse{
		Streams:    streams,
		Album:      album,
		Tracks:     tracks,
		TrackRanks: trackRanks,
	}

	json.NewEncoder(w).Encode(response)
}

func (c *StatsController) GetAlbumRankingsByURI(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userUUID, err := userOrFriendUUIDFromRequest(ctx, r)
	if err != nil {
		requests.RespondWithError(w, 401, fmt.Sprintf("parse user UUID: %s", err))
		return
	}
	code, spClient, err := client.ForUser(ctx, userUUID)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get client: %s", err), code)
		return
	}

	albumURI := mux.Vars(r)["spotify_uri"]
	albumID, err := service.IDFromURI(albumURI)
	if err != nil {
		http.Error(w, fmt.Sprintf("bad album uri: %s", err), http.StatusBadRequest)
		return
	}

	album, err := service.GetAlbum(ctx, spClient, albumID)
	if err != nil {
		http.Error(w, fmt.Sprintf("couldn't get album: %s", err), http.StatusInternalServerError)
		return
	}

	rankings := []TimeframeRanking{}

	transaction, err := db.Service().DB.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return
	}
	defer transaction.Commit()

	filter := getFilterParams(r)
	allRankings, responseCode, err := history.AlbumStreamRankingsByTimeframe(ctx, transaction, userUUID, filter, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), responseCode)
	}

	for _, timeframeRankings := range allRankings {
		for _, albumStreams := range timeframeRankings.AlbumStreams {
			if albumStreams.ID == string(album.ID) {
				ranking := TimeframeRanking{
					Position:             int(albumStreams.Rank),
					Timeframe:            timeframeRankings.Timeframe,
					StartDateUnixSeconds: timeframeRankings.StartDateUnixSeconds,
				}
				rankings = append(rankings, ranking)
			}
		}
	}

	json.NewEncoder(w).Encode(rankings)
}
